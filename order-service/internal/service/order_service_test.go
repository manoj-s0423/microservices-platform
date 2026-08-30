package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shopstream/order-service/internal/client"
	"github.com/shopstream/order-service/internal/models"
	"github.com/shopstream/order-service/internal/repository"
	"github.com/shopstream/order-service/internal/service"
)

// --- fakes ---

type fakeRepo struct {
	orders map[uuid.UUID]*models.Order
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{orders: map[uuid.UUID]*models.Order{}}
}

func (f *fakeRepo) Create(_ context.Context, order *models.Order) error {
	f.orders[order.ID] = order
	return nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Order, error) {
	o, ok := f.orders[id]
	if !ok {
		return nil, repository.ErrOrderNotFound
	}
	return o, nil
}

func (f *fakeRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]models.Order, error) {
	var result []models.Order
	for _, o := range f.orders {
		if o.UserID == userID {
			result = append(result, *o)
		}
	}
	return result, nil
}

func (f *fakeRepo) UpdateStatus(_ context.Context, id uuid.UUID, status models.OrderStatus) error {
	o, ok := f.orders[id]
	if !ok {
		return repository.ErrOrderNotFound
	}
	o.Status = status
	return nil
}

type fakeUsers struct {
	exists bool
}

func (f *fakeUsers) VerifyUserExists(_ context.Context, _ string) error {
	if !f.exists {
		return client.ErrUserNotFound
	}
	return nil
}

type fakeProducts struct {
	products map[string]*client.ProductInfo
}

func (f *fakeProducts) GetProduct(_ context.Context, productID string) (*client.ProductInfo, error) {
	p, ok := f.products[productID]
	if !ok {
		return nil, client.ErrProductNotFound
	}
	return p, nil
}

type fakePayments struct {
	result *client.ChargeResult
	err    error
}

func (f *fakePayments) Charge(_ context.Context, _ client.ChargeRequest) (*client.ChargeResult, error) {
	return f.result, f.err
}

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// --- tests ---

func TestCreateOrder_Success(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo()
	users := &fakeUsers{exists: true}
	products := &fakeProducts{products: map[string]*client.ProductInfo{
		productID.String(): {ID: productID.String(), SKU: "SKU-1", PriceCents: 1000, StockQuantity: 5, IsActive: true},
	}}
	payments := &fakePayments{result: &client.ChargeResult{Status: "SUCCEEDED", TransactionID: "txn_123"}}

	svc := service.NewOrderService(repo, users, products, payments, noopLogger())

	order, err := svc.CreateOrder(context.Background(), models.CreateOrderRequest{
		UserID: uuid.New().String(),
		Items:  []models.CreateOrderItemRequest{{ProductID: productID.String(), Quantity: 2}},
	})

	require.NoError(t, err)
	assert.Equal(t, models.OrderStatusConfirmed, order.Status)
	assert.Equal(t, 2000, order.TotalCents)
}

func TestCreateOrder_UnknownUser_ReturnsUserNotFound(t *testing.T) {
	repo := newFakeRepo()
	users := &fakeUsers{exists: false}
	products := &fakeProducts{products: map[string]*client.ProductInfo{}}
	payments := &fakePayments{}

	svc := service.NewOrderService(repo, users, products, payments, noopLogger())

	_, err := svc.CreateOrder(context.Background(), models.CreateOrderRequest{
		UserID: uuid.New().String(),
		Items:  []models.CreateOrderItemRequest{{ProductID: uuid.New().String(), Quantity: 1}},
	})

	assert.ErrorIs(t, err, service.ErrUserNotFound)
}

func TestCreateOrder_UnknownProduct_ReturnsProductNotFound(t *testing.T) {
	repo := newFakeRepo()
	users := &fakeUsers{exists: true}
	products := &fakeProducts{products: map[string]*client.ProductInfo{}}
	payments := &fakePayments{}

	svc := service.NewOrderService(repo, users, products, payments, noopLogger())

	_, err := svc.CreateOrder(context.Background(), models.CreateOrderRequest{
		UserID: uuid.New().String(),
		Items:  []models.CreateOrderItemRequest{{ProductID: uuid.New().String(), Quantity: 1}},
	})

	assert.ErrorIs(t, err, service.ErrProductNotFound)
}

func TestCreateOrder_InsufficientStock_ReturnsProductUnavailable(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo()
	users := &fakeUsers{exists: true}
	products := &fakeProducts{products: map[string]*client.ProductInfo{
		productID.String(): {ID: productID.String(), SKU: "SKU-1", PriceCents: 1000, StockQuantity: 1, IsActive: true},
	}}
	payments := &fakePayments{}

	svc := service.NewOrderService(repo, users, products, payments, noopLogger())

	_, err := svc.CreateOrder(context.Background(), models.CreateOrderRequest{
		UserID: uuid.New().String(),
		Items:  []models.CreateOrderItemRequest{{ProductID: productID.String(), Quantity: 5}},
	})

	assert.ErrorIs(t, err, service.ErrProductUnavailable)
}

func TestCreateOrder_PaymentDeclined_MarksOrderFailedButReturnsIt(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo()
	users := &fakeUsers{exists: true}
	products := &fakeProducts{products: map[string]*client.ProductInfo{
		productID.String(): {ID: productID.String(), SKU: "SKU-1", PriceCents: 1000, StockQuantity: 5, IsActive: true},
	}}
	payments := &fakePayments{result: &client.ChargeResult{Status: "DECLINED", Reason: "insufficient_funds"}}

	svc := service.NewOrderService(repo, users, products, payments, noopLogger())

	order, err := svc.CreateOrder(context.Background(), models.CreateOrderRequest{
		UserID: uuid.New().String(),
		Items:  []models.CreateOrderItemRequest{{ProductID: productID.String(), Quantity: 1}},
	})

	assert.ErrorIs(t, err, service.ErrPaymentDeclined)
	require.NotNil(t, order)
	assert.Equal(t, models.OrderStatusFailed, order.Status)
}

func TestCreateOrder_PaymentServiceUnreachable_MarksOrderFailed(t *testing.T) {
	productID := uuid.New()
	repo := newFakeRepo()
	users := &fakeUsers{exists: true}
	products := &fakeProducts{products: map[string]*client.ProductInfo{
		productID.String(): {ID: productID.String(), SKU: "SKU-1", PriceCents: 1000, StockQuantity: 5, IsActive: true},
	}}
	payments := &fakePayments{err: errors.New("connection refused")}

	svc := service.NewOrderService(repo, users, products, payments, noopLogger())

	order, err := svc.CreateOrder(context.Background(), models.CreateOrderRequest{
		UserID: uuid.New().String(),
		Items:  []models.CreateOrderItemRequest{{ProductID: productID.String(), Quantity: 1}},
	})

	assert.ErrorIs(t, err, service.ErrPaymentDeclined)
	assert.Equal(t, models.OrderStatusFailed, order.Status)
}

func TestGetOrder_NotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOrderService(repo, &fakeUsers{}, &fakeProducts{}, &fakePayments{}, noopLogger())

	_, err := svc.GetOrder(context.Background(), uuid.New())

	assert.ErrorIs(t, err, service.ErrOrderNotFound)
}
