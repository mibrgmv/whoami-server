package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gordle/gateway/internal/auth"
	"gordle/gateway/internal/config"
	"gordle/gateway/internal/middleware"
	"gordle/gateway/internal/models"
	"gordle/gateway/internal/repository"
	"gordle/libs/keycloak"
)

type AuthHandler struct {
	sessionRepo    repository.SessionRepository
	keycloakClient *keycloak.Client
	sessionConfig  config.SessionConfig
}

func NewAuthHandler(
	sessionRepo repository.SessionRepository,
	keycloakClient *keycloak.Client,
	sessionConfig config.SessionConfig,
) *AuthHandler {
	return &AuthHandler{
		sessionRepo:    sessionRepo,
		keycloakClient: keycloakClient,
		sessionConfig:  sessionConfig,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	h.authRedirect(c, "auth")
}

func (h *AuthHandler) Register(c *gin.Context) {
	h.authRedirect(c, "registrations")
}

func (h *AuthHandler) Action(c *gin.Context) {
	kcAction := c.Query("kc_action")
	if kcAction == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kc_action parameter required"})
		return
	}

	returnTo := sanitizeReturnTo(c.DefaultQuery("return_to", "/"))
	state := uuid.New().String()

	if err := h.sessionRepo.SetState(c.Request.Context(), state, returnTo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create state"})
		return
	}

	cfg := h.keycloakClient.Config()
	params := url.Values{
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {h.sessionConfig.CallbackURL},
		"response_type": {"code"},
		"scope":         {"openid"},
		"state":         {state},
		"kc_action":     {kcAction},
	}

	authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth?%s", browserBaseURL(cfg), cfg.Realm, params.Encode())
	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) authRedirect(c *gin.Context, endpoint string) {
	returnTo := sanitizeReturnTo(c.DefaultQuery("return_to", "/"))
	state := uuid.New().String()

	if err := h.sessionRepo.SetState(c.Request.Context(), state, returnTo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create state"})
		return
	}

	cfg := h.keycloakClient.Config()
	params := url.Values{
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {h.sessionConfig.CallbackURL},
		"response_type": {"code"},
		"scope":         {"openid"},
		"state":         {state},
	}

	authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/%s?%s", browserBaseURL(cfg), cfg.Realm, endpoint, params.Encode())
	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) Callback(c *gin.Context) {
	ctx := c.Request.Context()

	if errParam := c.Query("error"); errParam != "" {
		desc := c.Query("error_description")
		c.Redirect(http.StatusFound, "/?error="+url.QueryEscape(desc))
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code or state"})
		return
	}

	returnTo, err := h.sessionRepo.GetState(ctx, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired state"})
		return
	}

	kcActionStatus := c.Query("kc_action_status")

	tokenResp, err := h.keycloakClient.ExchangeAuthorizationCode(ctx, code, h.sessionConfig.CallbackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code"})
		return
	}

	claims, err := keycloak.ParseUnverifiedClaims(tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse token"})
		return
	}

	if kcActionStatus != "" {
		existingSessionID := h.getSessionIDFromCookie(c)
		if existingSessionID != "" {
			session, err := h.sessionRepo.Get(ctx, existingSessionID)
			if err == nil && session != nil {
				session.AccessToken = tokenResp.AccessToken
				session.IDToken = tokenResp.IDToken
				session.RefreshToken = tokenResp.RefreshToken
				session.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
				_ = h.sessionRepo.Update(ctx, existingSessionID, session)
				c.Redirect(http.StatusFound, returnTo)
				return
			}
		}
	}

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

	sessionData := &models.SessionData{
		AccessToken:   tokenResp.AccessToken,
		IDToken:       tokenResp.IDToken,
		RefreshToken:  tokenResp.RefreshToken,
		UserID:        claims.Subject,
		Username:      claims.PreferredUsername,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Roles:         strings.Join(roles, ","),
		ExpiresAt:     time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	sessionTTL := time.Duration(tokenResp.RefreshExpiresIn) * time.Second
	sessionID, err := h.sessionRepo.Create(ctx, sessionData, sessionTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	cookieMaxAge := tokenResp.RefreshExpiresIn
	if cookieMaxAge <= 0 {
		cookieMaxAge = 86400
	}
	c.SetCookie(h.cookieName(), sessionID, cookieMaxAge, "/", "", h.sessionConfig.CookieSecure, true)
	c.Redirect(http.StatusFound, returnTo)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	sessionID := h.getSessionIDFromCookie(c)

	var idToken string
	if sessionID != "" {
		session, err := h.sessionRepo.Get(ctx, sessionID)
		if err == nil && session != nil {
			idToken = session.IDToken
			_ = h.keycloakClient.RevokeToken(ctx, session.RefreshToken)
			_ = h.sessionRepo.Delete(ctx, sessionID)
		}
	}

	c.SetCookie(h.cookieName(), "", -1, "/", "", h.sessionConfig.CookieSecure, true)

	cfg := h.keycloakClient.Config()
	postLogoutURI := h.sessionConfig.CallbackURL
	if idx := strings.Index(postLogoutURI, "/api/"); idx != -1 {
		postLogoutURI = postLogoutURI[:idx]
	}

	params := url.Values{
		"client_id":                {cfg.ClientID},
		"post_logout_redirect_uri": {postLogoutURI},
	}
	if idToken != "" {
		params.Set("id_token_hint", idToken)
	}

	logoutURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout?%s",
		browserBaseURL(cfg), cfg.Realm, params.Encode())
	c.Redirect(http.StatusFound, logoutURL)
}

func (h *AuthHandler) Me(c *gin.Context) {
	ctx := c.Request.Context()

	userID, _ := ctx.Value(middleware.UserIDKey).(string)
	if userID == "" {
		c.JSON(http.StatusOK, gin.H{
			"username":        "",
			"email":           "",
			"isGuest":         false,
			"isAuthenticated": false,
		})
		return
	}

	isGuest, _ := ctx.Value(middleware.IsGuestKey).(bool)
	username, _ := ctx.Value(middleware.UsernameKey).(string)
	email, _ := ctx.Value(middleware.EmailKey).(string)

	c.JSON(http.StatusOK, gin.H{
		"username":        username,
		"email":           email,
		"isGuest":         isGuest,
		"isAuthenticated": true,
	})
}

func (h *AuthHandler) getSessionIDFromCookie(c *gin.Context) string {
	sessionID, err := c.Cookie(h.cookieName())
	if err != nil {
		return ""
	}
	return sessionID
}

func (h *AuthHandler) cookieName() string {
	if h.sessionConfig.CookieName != "" {
		return h.sessionConfig.CookieName
	}
	return "gordle_session"
}

func sanitizeReturnTo(returnTo string) string {
	if returnTo == "" || returnTo[0] != '/' || strings.HasPrefix(returnTo, "//") {
		return "/"
	}
	return returnTo
}

func browserBaseURL(cfg *keycloak.Config) string {
	if cfg.IssuerURL != "" {
		return cfg.IssuerURL
	}
	return cfg.BaseURL
}
