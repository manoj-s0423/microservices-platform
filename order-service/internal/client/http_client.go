// Package client holds resilient HTTP clients for calling sibling
// microservices (user-service, product-service, payment-service). All
// calls carry a bounded timeout and a small number of retries on transient
// failures, so a single flaky downstream call degrades gracefully instead
// of hanging the request indefinitely.
package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var (
	ErrDownstreamUnavailable = errors.New("downstream service unavailable")
	ErrDownstreamTimeout     = errors.New("downstream service timed out")
)

// ResilientClient wraps net/http with a timeout and retry policy shared by
// all downstream service clients.
type ResilientClient struct {
	httpClient    *http.Client
	retryAttempts int
	retryBackoff  time.Duration
	logger        *slog.Logger
	baseURL       string
	serviceName   string
}

func NewResilientClient(baseURL, serviceName string, timeout time.Duration, retryAttempts int, retryBackoff time.Duration, logger *slog.Logger) *ResilientClient {
	return &ResilientClient{
		httpClient:    &http.Client{Timeout: timeout},
		retryAttempts: retryAttempts,
		retryBackoff:  retryBackoff,
		logger:        logger,
		baseURL:       baseURL,
		serviceName:   serviceName,
	}
}

// Do executes an HTTP request built by reqFn (so a caller can freshly build
// the body/headers on every retry attempt) with retries on network errors,
// timeouts, and 502/503/504 responses.
func (c *ResilientClient) Do(ctx context.Context, method, path string, reqFn func() (*http.Request, error)) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retryAttempts; attempt++ {
		if attempt > 0 {
			backoff := c.retryBackoff * time.Duration(1<<uint(attempt-1)) // exponential backoff
			c.logger.Warn("retrying downstream call",
				"service", c.serviceName, "method", method, "path", path,
				"attempt", attempt, "backoff", backoff.String())
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := reqFn()
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = classifyErr(err)
			if !isRetryable(lastErr) {
				break
			}
			continue
		}

		if resp.StatusCode == http.StatusBadGateway ||
			resp.StatusCode == http.StatusServiceUnavailable ||
			resp.StatusCode == http.StatusGatewayTimeout {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("%w: status %d from %s", ErrDownstreamUnavailable, resp.StatusCode, c.serviceName)
			continue
		}

		return resp, nil
	}

	c.logger.Error("downstream call failed after retries",
		"service", c.serviceName, "method", method, "path", path, "error", lastErr)
	return nil, lastErr
}

func classifyErr(err error) error {
	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("%w: %v", ErrDownstreamTimeout, err)
	}
	return fmt.Errorf("%w: %v", ErrDownstreamUnavailable, err)
}

func isRetryable(err error) bool {
	return errors.Is(err, ErrDownstreamUnavailable) || errors.Is(err, ErrDownstreamTimeout)
}
