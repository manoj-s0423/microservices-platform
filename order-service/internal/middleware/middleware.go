// Package middleware provides cross-cutting Gin middleware: request-id
// propagation and structured request logging.
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-Id"

// RequestID reuses an inbound X-Request-Id (propagated from the API
// Gateway) or mints a new one, and echoes it back on the response so the
// same ID can be grepped across every service's logs for one request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("requestId", requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

// StructuredLogging logs one JSON line per request with method, path,
// status, latency, and the correlation id - the shape a log aggregator
// (CloudWatch/Loki/ELK) can query without regex parsing.
func StructuredLogging(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		requestID, _ := c.Get("requestId")
		logger.Info("request_completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"durationMs", duration.Milliseconds(),
			"requestId", requestID,
		)
	}
}
