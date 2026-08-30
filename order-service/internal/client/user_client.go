package client

import (
	"context"
	"fmt"
	"net/http"
)

var ErrUserNotFound = fmt.Errorf("user not found")

type UserClient struct {
	rc *ResilientClient
}

func NewUserClient(rc *ResilientClient) *UserClient {
	return &UserClient{rc: rc}
}

// VerifyUserExists calls user-service to confirm a userId is a real,
// active account before an order is placed against it.
func (c *UserClient) VerifyUserExists(ctx context.Context, userID string) error {
	url := fmt.Sprintf("%s/api/v1/users/%s", c.rc.baseURL, userID)

	resp, err := c.rc.Do(ctx, http.MethodGet, url, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrUserNotFound
	default:
		return fmt.Errorf("user-service returned unexpected status %d", resp.StatusCode)
	}
}
