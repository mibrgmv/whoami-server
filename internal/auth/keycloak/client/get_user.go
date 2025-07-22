package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GetUserResponse struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Enabled          bool   `json:"enabled"`
	EmailVerified    bool   `json:"emailVerified"`
	CreatedTimestamp int64  `json:"createdTimestamp"`
}

func (c *Client) GetUser(ctx context.Context, userID string) (*GetUserResponse, error) {
	adminToken, err := c.GetAdminToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin token: %w", err)
	}

	userURL := fmt.Sprintf("%s/admin/realms/%s/users/%s", c.config.BaseURL, c.config.Realm, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get user request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read get user response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("keycloak get user error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("failed to get user, status %d: %s", resp.StatusCode, string(body))
	}

	var userResp GetUserResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, fmt.Errorf("failed to parse get user response: %w", err)
	}

	return &userResp, nil
}
