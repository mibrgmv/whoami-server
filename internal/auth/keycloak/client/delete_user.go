package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DeleteUserResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func (c *Client) DeleteUser(ctx context.Context, userID string) (*DeleteUserResponse, error) {
	adminToken, err := c.GetAdminToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin token: %w", err)
	}

	userURL := fmt.Sprintf("%s/admin/realms/%s/users/%s", c.config.BaseURL, c.config.Realm, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create delete user request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("keycloak delete user error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("failed to delete user, status %d: %s", resp.StatusCode, string(body))
	}

	return &DeleteUserResponse{
		ID:      userID,
		Message: "User deleted successfully",
	}, nil
}
