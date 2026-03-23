package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type GuestTokenGenerator struct {
	secret []byte
}

func NewGuestTokenGenerator(secret string) *GuestTokenGenerator {
	return &GuestTokenGenerator{
		secret: []byte(secret),
	}
}

func (g *GuestTokenGenerator) GenerateToken() (string, string, error) {
	guestID := uuid.New().String()

	claims := GuestClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   guestID,
			Issuer:    GuestIssuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(GuestTokenTTL)),
		},
		Role: GuestRole,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(g.secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, guestID, nil
}
