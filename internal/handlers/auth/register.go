package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"whoami-server-gateway/internal/auth/keycloak/client"
)

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

	keycloakUser := client.CreateUserRequest{
		Username:      req.Username,
		Email:         req.Email,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Enabled:       true,
		EmailVerified: false,
		Credentials: []client.UserCredential{
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
