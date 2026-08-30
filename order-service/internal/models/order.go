package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusFailed    OrderStatus = "FAILED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type OrderItem struct {
	ID             uuid.UUID `json:"id"`
	ProductID      uuid.UUID `json:"productId"`
	SKU            string    `json:"sku"`
	Quantity       int       `json:"quantity"`
	UnitPriceCents int       `json:"unitPriceCents"`
}

type Order struct {
	ID         uuid.UUID   `json:"id"`
	UserID     uuid.UUID   `json:"userId"`
	Status     OrderStatus `json:"status"`
	TotalCents int         `json:"totalCents"`
	Currency   string      `json:"currency"`
	Items      []OrderItem `json:"items"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}

// CreateOrderItemRequest is the inbound shape for a single line item when
// placing an order; price is resolved server-side from product-service, it
// is never trusted from the client.
type CreateOrderItemRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	UserID string                   `json:"userId" binding:"required"`
	Items  []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}
