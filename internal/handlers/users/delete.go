package users

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

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
// @Router       /users/{id} [delete]
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
