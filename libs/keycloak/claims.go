package keycloak

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	PreferredUsername string                 `json:"preferred_username"`
	Email             string                 `json:"email"`
	EmailVerified     bool                   `json:"email_verified"`
	Name              string                 `json:"name"`
	GivenName         string                 `json:"given_name"`
	FamilyName        string                 `json:"family_name"`
	RealmAccess       map[string]interface{} `json:"realm_access"`
	ResourceAccess    map[string]interface{} `json:"resource_access"`
	Scope             string                 `json:"scope"`
	SessionState      string                 `json:"session_state"`
}

func (c *Claims) HasRole(role string) bool {
	if c == nil || c.RealmAccess == nil {
		return false
	}

	roles, ok := c.RealmAccess["roles"].([]interface{})
	if !ok {
		return false
	}

	for _, r := range roles {
		if str, ok := r.(string); ok && str == role {
			return true
		}
	}
	return false
}

func (c *Claims) IsAdmin() bool {
	return c.HasRole("admin")
}

func (c *Claims) Roles() []string {
	if c == nil || c.RealmAccess == nil {
		return nil
	}

	rolesInterface, ok := c.RealmAccess["roles"].([]interface{})
	if !ok {
		return nil
	}

	roles := make([]string, 0, len(rolesInterface))
	for _, r := range rolesInterface {
		if str, ok := r.(string); ok {
			roles = append(roles, str)
		}
	}
	return roles
}

func ParseUnverifiedClaims(accessToken string) (*Claims, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	return &claims, nil
}
