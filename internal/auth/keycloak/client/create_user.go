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
