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
