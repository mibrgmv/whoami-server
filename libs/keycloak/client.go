package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	config     *Config
	httpClient *http.Client
}

func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Config() *Config {
	return c.config
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorMessage     string `json:"errorMessage"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	IDToken          string `json:"id_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func (c *Client) tokenURL() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.config.BaseURL, c.config.Realm)
}

func (c *Client) postToken(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("keycloak error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("keycloak returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

func (c *Client) baseData() url.Values {
	data := url.Values{}
	data.Set("client_id", c.config.ClientID)
	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}
	return data
}

func (c *Client) ExchangeCredentialsForTokens(ctx context.Context, username, password string) (*TokenResponse, error) {
	data := c.baseData()
	data.Set("grant_type", "password")
	data.Set("username", username)
	data.Set("password", password)
	return c.postToken(ctx, data)
}

func (c *Client) ExchangeAuthorizationCode(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	data := c.baseData()
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	return c.postToken(ctx, data)
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := c.baseData()
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	return c.postToken(ctx, data)
}

func (c *Client) RevokeToken(ctx context.Context, refreshToken string) error {
	data := c.baseData()
	data.Set("token", refreshToken)
	data.Set("token_type_hint", "refresh_token")

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
