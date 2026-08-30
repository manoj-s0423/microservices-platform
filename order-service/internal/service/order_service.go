// Package service holds order-service's business logic: validating an
// order against user-service and product-service, persisting it, charging
// it via payment-service, and reconciling the final status.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/shopstream/order-service/internal/client"
	"github.com/shopstream/order-service/internal/models"
	"github.com/shopstream/order-service/internal/repository"
)

// Interfaces (rather than concrete *client.XClient types) so unit tests can
// substitute fakes with zero HTTP involved.
type UserVerifier interface {
	VerifyUserExists(ctx context.Context, userID string) error
}

type ProductFetcher interface {
	GetProduct(ctx context.Context, productID string) (*client.ProductInfo, error)
}

type PaymentCharger interface {
	Charge(ctx context.Context, req client.ChargeRequest) (*client.ChargeResult, error)
}

type OrderService struct {
	repo     repository.OrderRepository
	users    UserVerifier
	products ProductFetcher
	payments PaymentCharger
	logger   *slog.Logger
}

func NewOrderService(repo repository.OrderRepository, users UserVerifier, products ProductFetcher, payments PaymentCharger, logger *slog.Logger) *OrderService {
	return &OrderService{repo: repo, users: users, products: products, payments: payments, logger: logger}
}

func (s *OrderService) CreateOrder(ctx context.Context, req models.CreateOrderRequest) (*models.Order, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid userId", ErrValidation)
	}

	if err := s.users.VerifyUserExists(ctx, req.UserID); err != nil {
		if errors.Is(err, client.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}

	order := &models.Order{
		ID:       uuid.New(),
		UserID:   userID,
		Status:   models.OrderStatusPending,
		Currency: "USD",
	}

	total := 0
	for _, itemReq := range req.Items {
		product, err := s.products.GetProduct(ctx, itemReq.ProductID)
		if err != nil {
			if errors.Is(err, client.ErrProductNotFound) {
				return nil, fmt.Errorf("%w: product %s", ErrProductNotFound, itemReq.ProductID)
			}
			return nil, fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
		}
		if !product.IsActive || product.StockQuantity < itemReq.Quantity {
			return nil, fmt.Errorf("%w: product %s", ErrProductUnavailable, itemReq.ProductID)
		}

		productID, err := uuid.Parse(product.ID)
		if err != nil {
			return nil, fmt.Errorf("%w: product-service returned an invalid product id", ErrUpstreamUnavailable)
		}

		lineTotal := product.PriceCents * itemReq.Quantity
		total += lineTotal

		order.Items = append(order.Items, models.OrderItem{
			ID:             uuid.New(),
			ProductID:      productID,
			SKU:            product.SKU,
			Quantity:       itemReq.Quantity,
			UnitPriceCents: product.PriceCents,
		})
	}
	order.TotalCents = total

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("persisting order: %w", err)
	}

	chargeResult, err := s.payments.Charge(ctx, client.ChargeRequest{
		OrderID:     order.ID.String(),
		UserID:      order.UserID.String(),
		AmountCents: order.TotalCents,
		Currency:    order.Currency,
	})

	finalStatus := models.OrderStatusConfirmed
	if err != nil {
		s.logger.Error("payment charge failed, marking order as failed", "orderId", order.ID, "error", err)
		finalStatus = models.OrderStatusFailed
	} else if chargeResult.Status != "SUCCEEDED" {
		s.logger.Warn("payment declined", "orderId", order.ID, "reason", chargeResult.Reason)
		finalStatus = models.OrderStatusFailed
	}

	if updateErr := s.repo.UpdateStatus(ctx, order.ID, finalStatus); updateErr != nil {
		s.logger.Error("failed to persist final order status", "orderId", order.ID, "error", updateErr)
	}
	order.Status = finalStatus

	if finalStatus == models.OrderStatusFailed {
		return order, ErrPaymentDeclined
	}
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrOrderNotFound) {
		return nil, ErrOrderNotFound
	}
	return order, err
}

func (s *OrderService) ListOrdersByUser(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return s.repo.ListByUser(ctx, userID)
}
