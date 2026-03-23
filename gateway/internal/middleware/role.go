package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rolesVal := c.Request.Context().Value(RolesKey)
		rolesStr, ok := rolesVal.(string)
		if !ok || rolesStr == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "No roles found"})
			c.Abort()
			return
		}

		for _, r := range strings.Split(rolesStr, ",") {
			if r == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}
