package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gordle/gateway/internal/auth"
)

func GuestAuth(guestSecret string) gin.HandlerFunc {
	generator := auth.NewGuestTokenGenerator(guestSecret)

	return func(c *gin.Context) {
		token, guestID, err := generator.GenerateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate guest token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token": token,
			"guest_id":     guestID,
			"token_type":   "Bearer",
			"expires_in":   int(auth.GuestTokenTTL.Seconds()),
		})
	}
}
