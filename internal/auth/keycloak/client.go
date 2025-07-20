package keycloak

import (
	"bytes"
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
			Timeout: 30 * time.Second, // todo
		},
	}
}

func (c *Client) ExchangeCredentialsForTokens(ctx context.Context, username, password string) (*LoginResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.config.ClientID)
	data.Set("username", username)
	data.Set("password", password)

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	url := fmt.Sprintf("%s/admin/realms/%s/token", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
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

	var tokenResp LoginResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

func (c *Client) RefreshTokens(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.config.ClientID)
	data.Set("refresh_token", refreshToken)

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	url := fmt.Sprintf("%s/admin/realms/%s/token", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("keycloak refresh error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("keycloak refresh returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp LoginResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh token response: %w", err)
	}

	return &tokenResp, nil
}

// todo
func (c *Client) RevokeToken(ctx context.Context, refreshToken string) error {
	data := url.Values{}
	data.Set("client_id", c.config.ClientID)
	data.Set("token", refreshToken)
	data.Set("token_type_hint", "refresh_token")

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	url := fmt.Sprintf("%s/admin/realms/%s/logout", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
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

func (c *Client) GetAdminToken(ctx context.Context) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.config.AdminClientID)
	data.Set("client_secret", c.config.AdminClientSecret)

	url := fmt.Sprintf("%s/admin/realms/%s/token", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create admin token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get admin token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read admin token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get admin token, status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp LoginResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse admin token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

func (c *Client) CreateUser(ctx context.Context, userReq CreateUserRequest) (*CreateUserResponse, error) {
	adminToken, err := c.GetAdminToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin token: %w", err)
	}

	reqBody, err := json.Marshal(userReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user request: %w", err)
	}

	url := fmt.Sprintf("%s/admin/realms/%s/users", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read create user response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("keycloak user creation error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("failed to create user, status %d: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("no location header in response")
	}

	parts := strings.Split(location, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid location header format")
	}
	userID := parts[len(parts)-1]

	return &CreateUserResponse{
		ID:       userID,
		Username: userReq.Username,
		Email:    userReq.Email,
	}, nil
}
