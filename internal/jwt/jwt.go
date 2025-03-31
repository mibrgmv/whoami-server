package jwt

import (
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTKey = []byte("my_secret_key")

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string) (string, error) {
	_, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid UUID format: %w", err)
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTKey)
}

func ValidateToken(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTKey, nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid or expired token")
	}

	_, err = uuid.Parse(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("token contains invalid UUID: %w", err)
	}

	return claims.UserID, nil
}
