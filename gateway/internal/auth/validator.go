package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

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
