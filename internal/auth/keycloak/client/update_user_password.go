package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) UpdateUserPassword(ctx context.Context, userID, newPassword string) error {
	adminToken, err := c.GetAdminToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin token: %w", err)
	}

	passwordPayload := map[string]interface{}{
		"type":      "password",
		"value":     newPassword,
		"temporary": false,
	}

	reqBody, err := json.Marshal(passwordPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal password update request: %w", err)
	}

	passwordURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password", c.config.BaseURL, c.config.Realm, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, passwordURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create password update request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("user not found")
	}

	if resp.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "password policy") || strings.Contains(string(body), "policy") {
			return fmt.Errorf("password policy violation")
		}
		return fmt.Errorf("invalid password update request: %s", string(body))
	}

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return fmt.Errorf("keycloak password update error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return fmt.Errorf("failed to update password, status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
