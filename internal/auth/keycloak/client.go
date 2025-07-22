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

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func (c *Client) ExchangeCredentialsForTokens(ctx context.Context, username, password string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.config.ClientID)
	data.Set("username", username)
	data.Set("password", password)

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
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

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.config.ClientID)
	data.Set("refresh_token", refreshToken)

	if c.config.ClientSecret != "" {
		data.Set("client_secret", c.config.ClientSecret)
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
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

	var tokenResp TokenResponse
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

func (c *Client) GetAdminToken(ctx context.Context) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.config.AdminClientID)
	data.Set("client_secret", c.config.AdminClientSecret)

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
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

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse admin token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

type UserCredential struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

type CreateUserRequest struct {
	Username      string                 `json:"username"`
	Email         string                 `json:"email"`
	FirstName     string                 `json:"firstName,omitempty"`
	LastName      string                 `json:"lastName,omitempty"`
	Enabled       bool                   `json:"enabled"`
	EmailVerified bool                   `json:"emailVerified"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
	Credentials   []UserCredential       `json:"credentials,omitempty"`
}

type CreateUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
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

	userURL := fmt.Sprintf("%s/admin/realms/%s/users", c.config.BaseURL, c.config.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, userURL, bytes.NewReader(reqBody))
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

type UpdateUserRequest struct {
	ID        string `json:"id"`
	Username  string `json:"username,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

type UpdateUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (c *Client) UpdateUser(ctx context.Context, updateReq UpdateUserRequest) (*UpdateUserResponse, error) {
	adminToken, err := c.GetAdminToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin token: %w", err)
	}

	updatePayload := make(map[string]interface{})

	if updateReq.Username != "" {
		updatePayload["username"] = updateReq.Username
	}
	if updateReq.Email != "" {
		updatePayload["email"] = updateReq.Email
	}
	if updateReq.FirstName != "" {
		updatePayload["firstName"] = updateReq.FirstName
	}
	if updateReq.LastName != "" {
		updatePayload["lastName"] = updateReq.LastName
	}

	reqBody, err := json.Marshal(updatePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update request: %w", err)
	}

	userURL := fmt.Sprintf("%s/admin/realms/%s/users/%s", c.config.BaseURL, c.config.Realm, updateReq.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, userURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create update user request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read update user response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode == http.StatusConflict {
		if strings.Contains(string(body), "username") {
			return nil, fmt.Errorf("User exists with same username")
		}
		if strings.Contains(string(body), "email") {
			return nil, fmt.Errorf("User exists with same email")
		}
		return nil, fmt.Errorf("conflict updating user")
	}

	if resp.StatusCode != http.StatusNoContent {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("keycloak update user error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("failed to update user, status %d: %s", resp.StatusCode, string(body))
	}

	return &UpdateUserResponse{
		ID:       updateReq.ID,
		Username: updateReq.Username,
		Email:    updateReq.Email,
	}, nil
}

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

func (c *Client) VerifyUserPassword(ctx context.Context, userID, password string) error {
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	_, err = c.ExchangeCredentialsForTokens(ctx, user.Username, password)
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	return nil
}

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
