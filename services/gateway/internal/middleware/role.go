package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mibrgmv/whoami-server/shared/keycloak"
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

		if realmAccess, ok := keycloakClaims.RealmAccess["roles"].([]interface{}); ok {
			for _, r := range realmAccess {
				if roleStr, ok := r.(string); ok && roleStr == role {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
