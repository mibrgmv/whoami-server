package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	GuestIssuer   = "gordle-guest"
	GuestRole     = "guest"
	UserRole      = "user"
	AdminRole     = "admin"
	GuestTokenTTL = 24 * time.Hour
)

type GuestClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}
