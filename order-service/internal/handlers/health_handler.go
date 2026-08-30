package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool *pgxpool.Pool
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool}
}

// Health is liveness only - no dependency checks, so an orchestrator never
// restarts this pod solely because PostgreSQL is briefly unreachable.
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "UP",
		"service":   "order-service",
		"timestamp": time.Now().UTC(),
	})
}

// Ready pings the database with a short timeout; used for the k8s
// readiness probe so traffic is pulled from a pod that can't reach Postgres.
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1500*time.Millisecond)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":       "DEGRADED",
			"dependencies": gin.H{"database": "DOWN"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "UP",
		"dependencies": gin.H{"database": "UP"},
	})
}
