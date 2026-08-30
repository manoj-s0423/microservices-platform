package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ChargeRequest struct {
	OrderID     string `json:"orderId"`
	UserID      string `json:"userId"`
	AmountCents int    `json:"amountCents"`
	Currency    string `json:"currency"`
}

type ChargeResult struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"` // SUCCEEDED | FAILED | DECLINED
	Reason        string `json:"reason,omitempty"`
}

type PaymentClient struct {
	rc *ResilientClient
}

func NewPaymentClient(rc *ResilientClient) *PaymentClient {
	return &PaymentClient{rc: rc}
}

// Charge sends orderId as an Idempotency-Key so that a retried request
// (from the ResilientClient's own retry policy, or a client-side retry
// after a timeout) is safe: payment-service treats a repeated key as
// "return the original result" rather than charging twice.
func (c *PaymentClient) Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
	url := fmt.Sprintf("%s/api/v1/payments", c.rc.baseURL)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.rc.Do(ctx, http.MethodPost, url, func() (*http.Request, error) {
		r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", req.OrderID)
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusPaymentRequired {
		return nil, fmt.Errorf("payment-service returned unexpected status %d", resp.StatusCode)
	}

	var result ChargeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding payment-service response: %w", err)
	}
	return &result, nil
}
