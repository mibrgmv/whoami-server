package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"whoami-server-gateway/internal/auth/keycloak"
	"whoami-server-gateway/protogen/golang/user"
)

type Handler struct {
	keycloak *keycloak.Client
	user     *user.UserServiceClient
}

func NewHandler(keycloak *keycloak.Client, user *user.UserServiceClient) *Handler {
	return &Handler{
		keycloak: keycloak,
		user:     user,
	}
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	tokens, err := h.keycloak.ExchangeCredentialsForTokens(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.keycloak.RevokeToken(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.keycloak.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Username  string `json:"username" binding:"required,min=3,max=50"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=8"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	keycloakUser := keycloak.CreateUserRequest{
		Username:      req.Username,
		Email:         req.Email,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Enabled:       true,
		EmailVerified: false,
		Credentials: []keycloak.UserCredential{
			{
				Type:      "password",
				Value:     req.Password,
				Temporary: false,
			},
		},
	}

	keycloakResp, err := h.keycloak.CreateUser(c.Request.Context(), keycloakUser)
	if err != nil {
		if strings.Contains(err.Error(), "User exists with same username") {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
			return
		}
		if strings.Contains(err.Error(), "User exists with same email") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ID":       keycloakResp.ID,
		"Username": keycloakResp.Username,
		"Email":    keycloakResp.Email,
		"Message":  "User created successfully",
	})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	targetUserID := c.Param("id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if authUserID.(string) != targetUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this user"})
		return
	}

	deletedUser, err := h.keycloak.DeleteUser(c.Request.Context(), targetUserID)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, deletedUser)
}
