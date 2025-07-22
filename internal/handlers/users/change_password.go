package users

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

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
// @Router       /users/{id}/password [put]
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
