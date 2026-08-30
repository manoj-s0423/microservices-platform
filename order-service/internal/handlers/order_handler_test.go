package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shopstream/order-service/internal/client"
	"github.com/shopstream/order-service/internal/handlers"
	"github.com/shopstream/order-service/internal/models"
	"github.com/shopstream/order-service/internal/repository"
	"github.com/shopstream/order-service/internal/service"
)

type stubRepo struct{ orders map[uuid.UUID]*models.Order }

func (r *stubRepo) Create(_ context.Context, o *models.Order) error { r.orders[o.ID] = o; return nil }
func (r *stubRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Order, error) {
	if o, ok := r.orders[id]; ok {
		return o, nil
	}
	return nil, repository.ErrOrderNotFound
}
func (r *stubRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]models.Order, error) {
	var out []models.Order
	for _, o := range r.orders {
		if o.UserID == userID {
			out = append(out, *o)
		}
	}
	return out, nil
}
func (r *stubRepo) UpdateStatus(_ context.Context, id uuid.UUID, s models.OrderStatus) error {
	if o, ok := r.orders[id]; ok {
		o.Status = s
		return nil
	}
	return repository.ErrOrderNotFound
}

type stubUsers struct{}

func (stubUsers) VerifyUserExists(context.Context, string) error { return nil }

type stubProducts struct{ productID string }

func (s stubProducts) GetProduct(_ context.Context, id string) (*client.ProductInfo, error) {
	if id != s.productID {
		return nil, client.ErrProductNotFound
	}
	return &client.ProductInfo{ID: id, SKU: "SKU-1", PriceCents: 500, StockQuantity: 10, IsActive: true}, nil
}

type stubPayments struct{}

func (stubPayments) Charge(context.Context, client.ChargeRequest) (*client.ChargeResult, error) {
	return &client.ChargeResult{Status: "SUCCEEDED", TransactionID: "txn_test"}, nil
}

func setupRouter() (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	productID := uuid.New().String()
	repo := &stubRepo{orders: map[uuid.UUID]*models.Order{}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.NewOrderService(repo, stubUsers{}, stubProducts{productID: productID}, stubPayments{}, logger)
	h := handlers.NewOrderHandler(svc)

	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.POST("/orders", h.CreateOrder)
	v1.GET("/orders/:id", h.GetOrder)
	v1.GET("/orders", h.ListOrders)

	return router, productID
}

func TestCreateOrder_ReturnsCreated(t *testing.T) {
	router, productID := setupRouter()

	body, _ := json.Marshal(models.CreateOrderRequest{
		UserID: uuid.New().String(),
		Items:  []models.CreateOrderItemRequest{{ProductID: productID, Quantity: 1}},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp models.Order
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, models.OrderStatusConfirmed, resp.Status)
}

func TestCreateOrder_MissingItems_ReturnsBadRequest(t *testing.T) {
	router, _ := setupRouter()

	body := []byte(`{"userId": "` + uuid.New().String() + `", "items": []}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetOrder_NotFound_Returns404(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetOrder_InvalidUUID_Returns400(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/not-a-uuid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
