package users

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"whoami-server-gateway/internal/auth/keycloak/client"
)

type UpdateUserRequest struct {
	Username  string `json:"username,omitempty" binding:"omitempty,min=3,max=50" example:"johndoe_updated"`
	Email     string `json:"email,omitempty" binding:"omitempty,email" example:"john.updated@example.com"`
	FirstName string `json:"first_name,omitempty" example:"John"`
	LastName  string `json:"last_name,omitempty" example:"Doe"`
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
// @Router       /users/{id} [put]
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

	updateReq := client.UpdateUserRequest{
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
