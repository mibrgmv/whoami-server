package user

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"whoami-server/cmd/gateway/internal/models"
	"whoami-server/cmd/gateway/internal/services/user"
	"whoami-server/internal/jwt"
)

type Handler struct {
	userService *user.Service
}

func NewHandler(userService *user.Service) *Handler {
	return &Handler{
		userService: userService,
	}
}

type TokenResponse struct {
	Token  string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID int64  `json:"user_id" example:"1"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with the provided username and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "User registration details"
// @Success 201 {object} models.User "User created successfully"
// @Failure 400
// @Failure 409
// @Failure 500
// @Router /register [post]
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	userResponse, err := h.userService.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	userResponse.Password = ""

	c.JSON(http.StatusCreated, userResponse)
}

// Login godoc
// @Summary User login
// @Description Authenticate a user and return a JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "User login credentials"
// @Success 200 {object} TokenResponse "Successfully logged in with token and user ID"
// @Failure 400
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	userID, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, user.ErrUserNotFound) {
			status = http.StatusNotFound
		}

		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	tokenString, err := jwt.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   tokenString,
		"user_id": userID,
	})
}

// GetCurrent godoc
// @Summary Get current user profile
// @Description Retrieve the profile of the currently authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User "User profile"
// @Failure 401
// @Failure 404
// @Failure 500
// @Router /users/current [get]
func (h *Handler) GetCurrent(c *gin.Context) {
	userID, exists := jwt.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userResponse, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Printf("Error retrieving user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	userResponse.Password = ""

	c.JSON(http.StatusOK, userResponse)
}

// GetAll godoc
// @Summary Get all users
// @Description Retrieve a list of all registered users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User "List of users"
// @Failure 401
// @Failure 500
// @Router /users [get]
func (h *Handler) GetAll(c *gin.Context) {
	_, exists := jwt.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	users, err := h.userService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, users)
}
