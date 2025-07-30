package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"whoami-server-gateway/internal/auth/keycloak"
)

type Handler struct {
	keycloak *keycloak.Client
}

func NewHandler(keycloak *keycloak.Client) *Handler {
	return &Handler{
		keycloak: keycloak,
	}
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid credentials"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int    `json:"expires_in" example:"3600"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user with username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
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

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50" example:"johndoe"`
	Email     string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
	FirstName string `json:"first_name,omitempty" example:"John"`
	LastName  string `json:"last_name,omitempty" example:"Doe"`
}

type RegisterSuccessResponse struct {
	ID       string `json:"ID" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`
	Username string `json:"Username" example:"johndoe"`
	Email    string `json:"Email" example:"john.doe@example.com"`
	Message  string `json:"Message" example:"User created successfully"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account through Keycloak identity provider
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "User registration details"
// @Success      201  {object}  RegisterSuccessResponse "User created successfully"
// @Failure      400  {object}  ErrorResponse "Bad request - Invalid input data"
// @Failure      409  {object}  ErrorResponse "Conflict - Username or email already exists"
// @Failure      500  {object}  ErrorResponse "Internal server error - Failed to create user"
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

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

	c.JSON(http.StatusCreated, RegisterSuccessResponse{
		ID:       keycloakResp.ID,
		Username: keycloakResp.Username,
		Email:    keycloakResp.Email,
		Message:  "User created successfully",
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int    `json:"expires_in" example:"3600"`
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Generate new access and refresh tokens using a valid refresh token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RefreshTokenRequest true "Refresh token details"
// @Success      200  {object}  RefreshTokenResponse "Tokens refreshed successfully"
// @Failure      400  {object}  ErrorResponse "Bad request - Invalid input data"
// @Failure      401  {object}  ErrorResponse "Unauthorized - Invalid refresh token"
// @Router       /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	tokens, err := h.keycloak.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	response := RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}

	c.JSON(http.StatusOK, response)
}

// todo structs and docs
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
