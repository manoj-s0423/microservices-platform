package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ProductInfo struct {
	ID            string `json:"id"`
	SKU           string `json:"sku"`
	Name          string `json:"name"`
	PriceCents    int    `json:"price_cents"`
	StockQuantity int    `json:"stock_quantity"`
	IsActive      bool   `json:"is_active"`
}

var ErrProductNotFound = fmt.Errorf("product not found")

type ProductClient struct {
	rc *ResilientClient
}

func NewProductClient(rc *ResilientClient) *ProductClient {
	return &ProductClient{rc: rc}
}

func (c *ProductClient) GetProduct(ctx context.Context, productID string) (*ProductInfo, error) {
	url := fmt.Sprintf("%s/api/v1/products/%s", c.rc.baseURL, productID)

	resp, err := c.rc.Do(ctx, http.MethodGet, url, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrProductNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product-service returned unexpected status %d", resp.StatusCode)
	}

	var info ProductInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding product-service response: %w", err)
	}
	return &info, nil
}
