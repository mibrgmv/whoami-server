package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

type GuestTokenValidator struct {
	secret []byte
}

func NewGuestTokenValidator(secret string) *GuestTokenValidator {
	return &GuestTokenValidator{
		secret: []byte(secret),
	}
}

func (v *GuestTokenValidator) ValidateToken(tokenString string) (*GuestClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &GuestClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*GuestClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	if claims.Issuer != GuestIssuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	return claims, nil
}

// IsGuestToken checks if a token string is a guest token by checking the issuer
func IsGuestToken(tokenString string) bool {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &GuestClaims{})
	if err != nil {
		return false
	}

	claims, ok := token.Claims.(*GuestClaims)
	if !ok {
		return false
	}

	return claims.Issuer == GuestIssuer
}
