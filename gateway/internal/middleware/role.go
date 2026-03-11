package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gordle/libs/keycloak"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Claims not found"})
			c.Abort()
			return
		}

		keycloakClaims, ok := claims.(*keycloak.Claims)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		if !keycloakClaims.HasRole(role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}
