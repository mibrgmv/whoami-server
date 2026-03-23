package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"

	"gordle/gateway/internal/auth"
	"gordle/gateway/internal/models"
	"gordle/gateway/internal/repository"
	"gordle/libs/keycloak"
)

type contextKey string

const (
	UserIDKey        contextKey = "user_id"
	UsernameKey      contextKey = "username"
	EmailKey         contextKey = "email"
	EmailVerifiedKey contextKey = "email_verified"
	RolesKey         contextKey = "roles"
	IsGuestKey       contextKey = "is_guest"
)

type AuthFailureRecorder interface {
	RecordAuthFailure(reason string)
}

type AuthConfig struct {
	KeycloakValidator *auth.KeycloakValidator
	GuestValidator    *auth.GuestTokenValidator
	SessionRepo       repository.SessionRepository
	KeycloakClient    *keycloak.Client
	CookieName        string
	Metrics           AuthFailureRecorder
}

type authMiddleware struct {
	config       AuthConfig
	refreshGroup singleflight.Group
}

func Auth(cfg AuthConfig) gin.HandlerFunc {
	m := &authMiddleware{config: cfg}
	return m.required
}

func AuthOptional(cfg AuthConfig) gin.HandlerFunc {
	m := &authMiddleware{config: cfg}
	return m.optional
}

func (m *authMiddleware) cookieName() string {
	if m.config.CookieName != "" {
		return m.config.CookieName
	}
	return "gordle_session"
}

func (m *authMiddleware) recordFailure(reason string) {
	if m.config.Metrics != nil {
		m.config.Metrics.RecordAuthFailure(reason)
	}
}

func (m *authMiddleware) required(c *gin.Context) {
	if m.resolveBearer(c) {
		c.Next()
		return
	}

	if m.resolveSession(c) {
		c.Next()
		return
	}

	m.recordFailure("missing_token")
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
	c.Abort()
}

func (m *authMiddleware) optional(c *gin.Context) {
	if m.resolveBearer(c) {
		c.Next()
		return
	}

	m.resolveSession(c)
	c.Next()
}

func (m *authMiddleware) resolveBearer(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return false
	}

	if m.config.GuestValidator != nil && auth.IsGuestToken(tokenString) {
		guestClaims, err := m.config.GuestValidator.ValidateToken(tokenString)
		if err != nil {
			m.recordFailure("invalid_guest_token")
			return false
		}
		m.setGuestContext(c, guestClaims)
		return true
	}

	if m.config.KeycloakValidator != nil {
		claims, err := m.config.KeycloakValidator.ValidateToken(tokenString)
		if err != nil {
			m.recordFailure("invalid_token")
			return false
		}
		m.setKeycloakContext(c, claims)
		return true
	}

	return false
}

func (m *authMiddleware) resolveSession(c *gin.Context) bool {
	if m.config.SessionRepo == nil {
		return false
	}

	sessionID, err := c.Cookie(m.cookieName())
	if err != nil || sessionID == "" {
		return false
	}

	ctx := c.Request.Context()
	session, err := m.config.SessionRepo.Get(ctx, sessionID)
	if err != nil || session == nil {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		if m.config.KeycloakClient == nil {
			return false
		}

		result, err, _ := m.refreshGroup.Do(sessionID, func() (interface{}, error) {
			return m.config.KeycloakClient.RefreshToken(ctx, session.RefreshToken)
		})
		if err != nil {
			return false
		}

		tokenResp := result.(*keycloak.TokenResponse)
		session.AccessToken = tokenResp.AccessToken
		session.IDToken = tokenResp.IDToken
		session.RefreshToken = tokenResp.RefreshToken
		session.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

		_ = m.config.SessionRepo.Update(ctx, sessionID, session)
	}

	m.setSessionContext(c, session)
	c.Request.Header.Set("Authorization", "Bearer "+session.AccessToken)

	return true
}

func (m *authMiddleware) setKeycloakContext(c *gin.Context, claims *keycloak.Claims) {
	roles := claims.Roles()
	hasUserRole := false
	for _, r := range roles {
		if r == auth.UserRole {
			hasUserRole = true
			break
		}
	}
	if !hasUserRole {
		roles = append(roles, auth.UserRole)
	}
	rolesStr := strings.Join(roles, ",")

	setContext(c, claims.Subject, claims.PreferredUsername, claims.Email, claims.EmailVerified, rolesStr, false)
}

func (m *authMiddleware) setGuestContext(c *gin.Context, claims *auth.GuestClaims) {
	setContext(c, claims.Subject, "", "", false, claims.Role, true)
}

func (m *authMiddleware) setSessionContext(c *gin.Context, session *models.SessionData) {
	setContext(c, session.UserID, session.Username, session.Email, session.EmailVerified, session.Roles, false)
}

func setContext(c *gin.Context, userID, username, email string, emailVerified bool, roles string, isGuest bool) {
	c.Set("user_id", userID)
	c.Set("username", username)
	c.Set("email", email)
	c.Set("email_verified", emailVerified)

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, UsernameKey, username)
	ctx = context.WithValue(ctx, EmailKey, email)
	ctx = context.WithValue(ctx, EmailVerifiedKey, emailVerified)
	ctx = context.WithValue(ctx, RolesKey, roles)
	ctx = context.WithValue(ctx, IsGuestKey, isGuest)
	c.Request = c.Request.WithContext(ctx)
}
