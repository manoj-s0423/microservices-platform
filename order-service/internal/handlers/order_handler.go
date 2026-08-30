package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/shopstream/order-service/internal/models"
	"github.com/shopstream/order-service/internal/service"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": err.Error()})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), req)
	if err != nil {
		writeServiceError(c, err, order)
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "order id must be a UUID"})
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err, nil)
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userIDParam := c.Query("userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id", "message": "userId query param must be a UUID"})
		return
	}

	orders, err := h.orderService.ListOrdersByUser(c.Request.Context(), userID)
	if err != nil {
		writeServiceError(c, err, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": orders, "total": len(orders)})
}

// writeServiceError maps domain errors to HTTP status codes. When order is
// non-nil (order creation partially succeeded but payment failed), it is
// included in the error body so the client can see the FAILED order it
// still owns rather than just an opaque error.
func writeServiceError(c *gin.Context, err error, order *models.Order) {
	body := gin.H{"error": "internal_error", "message": err.Error()}
	if order != nil {
		body["order"] = order
	}

	switch {
	case errors.Is(err, service.ErrValidation):
		body["error"] = "validation_error"
		c.JSON(http.StatusBadRequest, body)
	case errors.Is(err, service.ErrUserNotFound):
		body["error"] = "user_not_found"
		c.JSON(http.StatusNotFound, body)
	case errors.Is(err, service.ErrProductNotFound):
		body["error"] = "product_not_found"
		c.JSON(http.StatusNotFound, body)
	case errors.Is(err, service.ErrProductUnavailable):
		body["error"] = "product_unavailable"
		c.JSON(http.StatusConflict, body)
	case errors.Is(err, service.ErrOrderNotFound):
		body["error"] = "order_not_found"
		c.JSON(http.StatusNotFound, body)
	case errors.Is(err, service.ErrPaymentDeclined):
		body["error"] = "payment_declined"
		c.JSON(http.StatusPaymentRequired, body)
	case errors.Is(err, service.ErrUpstreamUnavailable):
		body["error"] = "upstream_unavailable"
		c.JSON(http.StatusBadGateway, body)
	default:
		c.JSON(http.StatusInternalServerError, body)
	}
}
