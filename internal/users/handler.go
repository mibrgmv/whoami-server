package users

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
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

type UpdateUserRequest struct {
	Username  string `json:"username,omitempty" binding:"omitempty,min=3,max=50" example:"johndoe_updated"`
	Email     string `json:"email,omitempty" binding:"omitempty,email" example:"john.updated@example.com"`
	FirstName string `json:"firstName,omitempty" example:"John"`
	LastName  string `json:"lastName,omitempty" example:"Doe"`
}

type UpdateUserResponse struct {
	ID       string `json:"id" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`
	Username string `json:"username" example:"johndoe_updated"`
	Email    string `json:"email" example:"john.updated@example.com"`
	Message  string `json:"message" example:"User updated successfully"`
}

// UpdateUser godoc
// @Summary      Update user profile
// @Description  Update the authenticated user's profile information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id      path  string             true  "User ID"
// @Param        request body  UpdateUserRequest  true  "User update details"
// @Success      200     {object}  UpdateUserResponse "User updated successfully"
// @Failure      400     {object}  ErrorResponse "Bad request - Invalid input data or User ID is required"
// @Failure      401     {object}  ErrorResponse "Unauthorized - User not authenticated"
// @Failure      403     {object}  ErrorResponse "Forbidden - Not authorized to update this user"
// @Failure      404     {object}  ErrorResponse "Not found - User not found"
// @Failure      409     {object}  ErrorResponse "Conflict - Username or email already exists"
// @Failure      500     {object}  ErrorResponse "Internal server error - Failed to update user"
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	targetUserID := c.Param("id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "User ID is required"})
		return
	}

	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	if authUserID.(string) != targetUserID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "You are not authorized to update this user"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	updateReq := keycloak.UpdateUserRequest{
		ID: targetUserID,
	}

	if req.Username != "" {
		updateReq.Username = req.Username
	}
	if req.Email != "" {
		updateReq.Email = req.Email
	}
	if req.FirstName != "" {
		updateReq.FirstName = req.FirstName
	}
	if req.LastName != "" {
		updateReq.LastName = req.LastName
	}

	updatedUser, err := h.keycloak.UpdateUser(c.Request.Context(), updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		if strings.Contains(err.Error(), "User exists with same username") {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Username already exists"})
			return
		}
		if strings.Contains(err.Error(), "User exists with same email") {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Email already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update user"})
		return
	}

	response := UpdateUserResponse{
		ID:       updatedUser.ID,
		Username: updatedUser.Username,
		Email:    updatedUser.Email,
		Message:  "User updated successfully",
	}

	c.JSON(http.StatusOK, response)
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"currentPass123"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"newSecurePass456"`
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"newSecurePass456"`
}

type ChangePasswordResponse struct {
	Message string `json:"message" example:"Password updated successfully"`
}

// ChangePassword godoc
// @Summary      Change user password
// @Description  Change the authenticated user's password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id      path  string                 true  "User ID"
// @Param        request body  ChangePasswordRequest  true  "Password change details"
// @Success      200     {object}  ChangePasswordResponse "Password updated successfully"
// @Failure      400     {object}  ErrorResponse "Bad request - Invalid input data or password policy violation"
// @Failure      401     {object}  ErrorResponse "Unauthorized - User not authenticated or incorrect current password"
// @Failure      403     {object}  ErrorResponse "Forbidden - Not authorized to change this user's password"
// @Failure      404     {object}  ErrorResponse "Not found - User not found"
// @Failure      500     {object}  ErrorResponse "Internal server error - Failed to update password"
// @Security     BearerAuth
// @Router       /api/v1/users/{id}/password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	targetUserID := c.Param("id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "User ID is required"})
		return
	}

	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	if authUserID.(string) != targetUserID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "You are not authorized to change this user's password"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "New password and confirmation do not match"})
		return
	}

	if req.CurrentPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "New password must be different from current password"})
		return
	}

	err := h.keycloak.VerifyUserPassword(c.Request.Context(), targetUserID, req.CurrentPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Current password is incorrect"})
		return
	}

	err = h.keycloak.UpdateUserPassword(c.Request.Context(), targetUserID, req.NewPassword)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		if strings.Contains(err.Error(), "password policy") {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Password does not meet policy requirements"})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, ChangePasswordResponse{Message: "Password updated successfully"})
}

type DeleteUserResponse struct {
	ID      string `json:"id" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`
	Message string `json:"message" example:"User deleted successfully"`
}

// DeleteUser godoc
// @Summary      Delete user account
// @Description  Delete the authenticated user's account from Keycloak
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  DeleteUserResponse "User deleted successfully"
// @Failure      400  {object}  ErrorResponse "Bad request - User ID is required"
// @Failure      401  {object}  ErrorResponse "Unauthorized - User not authenticated"
// @Failure      403  {object}  ErrorResponse "Forbidden - Not authorized to delete this user"
// @Failure      404  {object}  ErrorResponse "Not found - User not found"
// @Failure      500  {object}  ErrorResponse "Internal server error - Failed to delete user"
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	targetUserID := c.Param("id")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "User ID is required"})
		return
	}

	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	if authUserID.(string) != targetUserID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "You are not authorized to delete this user"})
		return
	}

	deletedUser, err := h.keycloak.DeleteUser(c.Request.Context(), targetUserID)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete user"})
		return
	}

	response := DeleteUserResponse{
		ID:      deletedUser.ID,
		Message: "User deleted successfully",
	}

	c.JSON(http.StatusOK, response)
}

type GetCurrentUserResponse struct {
	ID            string `json:"id" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`
	Username      string `json:"username" example:"john_doe"`
	Email         string `json:"email" example:"john.doe@example.com"`
	FirstName     string `json:"firstName" example:"John"`
	LastName      string `json:"lastName" example:"Doe"`
	Enabled       bool   `json:"enabled" example:"true"`
	EmailVerified bool   `json:"emailVerified" example:"true"`
	CreatedAt     string `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// GetCurrentUser godoc
// @Summary      Get current user profile
// @Description  Get the authenticated user's profile information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200  {object}  GetCurrentUserResponse "User profile retrieved successfully"
// @Failure      401  {object}  ErrorResponse "Unauthorized - User not authenticated"
// @Failure      404  {object}  ErrorResponse "Not found - User not found"
// @Failure      500  {object}  ErrorResponse "Internal server error - Failed to retrieve user"
// @Security     BearerAuth
// @Router       /api/v1/users/current [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID := authUserID.(string)

	kcUser, err := h.keycloak.GetUser(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve user"})
		return
	}

	response := GetCurrentUserResponse{
		ID:            kcUser.ID,
		Username:      kcUser.Username,
		Email:         kcUser.Email,
		FirstName:     kcUser.FirstName,
		LastName:      kcUser.LastName,
		Enabled:       kcUser.Enabled,
		EmailVerified: kcUser.EmailVerified,
		CreatedAt:     time.Unix(kcUser.CreatedTimestamp/1000, 0).UTC().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

type BatchGetUsersRequest struct {
	PageSize int32 `form:"page_size,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
	Offset   int32 `form:"offset,omitempty" binding:"min=0" example:"0"`
}

type User struct {
	ID        string `json:"id" example:"f47ac10b-58cc-4372-a567-0e02b2c3d479"`
	Username  string `json:"username" example:"johndoe"`
	Email     string `json:"email" example:"john@example.com"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

type BatchGetUsersResponse struct {
	Users      []User `json:"users"`
	NextOffset *int32 `json:"next_offset,omitempty" example:"20"`
}

// BatchGetUsers godoc
// @Summary      Get users with pagination
// @Description  Retrieve a paginated list of users using offset-based pagination
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        page_size   query  int     false  "Number of users per page (1-100, default: 10)"  minimum(1)  maximum(100)  default(10)
// @Param        offset      query  int     false  "Starting position for pagination (default: 0)"  minimum(0)  default(0)
// @Success      200         {object}  BatchGetUsersResponse "Users retrieved successfully with next_offset for continuation"
// @Failure      400         {object}  ErrorResponse "Bad request - Invalid pagination parameters or negative offset"
// @Failure      401         {object}  ErrorResponse "Unauthorized - User not authenticated"
// @Failure      403         {object}  ErrorResponse "Forbidden - Insufficient permissions to list users"
// @Failure      500         {object}  ErrorResponse "Internal server error - Failed to retrieve users"
// @Security     BearerAuth
// @Router       /api/v1/users [get]
func (h *Handler) BatchGetUsers(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req BatchGetUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if req.Offset < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Offset must be non-negative"})
		return
	}

	// todo cache
	keycloakReq := keycloak.BatchGetUsersRequest{
		PageSize: req.PageSize,
		First:    req.Offset,
	}

	keycloakResp, err := h.keycloak.BatchGetUsers(c.Request.Context(), keycloakReq)
	if err != nil {
		if strings.Contains(err.Error(), "invalid page token") {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid page token"})
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "Insufficient permissions to list users"})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to retrieve users"})
		return
	}

	users := make([]User, len(keycloakResp.Users))
	for i, kcUser := range keycloakResp.Users {
		users[i] = User{
			ID:        kcUser.ID,
			Username:  kcUser.Username,
			Email:     kcUser.Email,
			FirstName: kcUser.FirstName,
			LastName:  kcUser.LastName,
			CreatedAt: time.Unix(kcUser.CreatedTimestamp/1000, 0).UTC().Format(time.RFC3339),
		}
	}

	response := BatchGetUsersResponse{
		Users:      users,
		NextOffset: keycloakResp.NextFirst,
	}

	c.JSON(http.StatusOK, response)
}
