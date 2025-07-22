package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// todo
func (c *Client) RevokeToken(ctx context.Context, refreshToken string) error {
	data := url.Values{}
	data.Set("client_id", c.config.ClientID)
	data.Set("token", refreshToken)
	data.Set("token_type_hint", "refresh_token")

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	revokeURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make revoke request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to revoke token, status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
