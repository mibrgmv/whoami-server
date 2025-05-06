package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var (
	AccessTokenKey  = []byte("access_secret_key")
	RefreshTokenKey = []byte("refresh_secret_key")
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    string    `json:"user_id"`
	TokenID   string    `json:"token_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateTokenPair(userID string) (accessToken string, refreshToken string, err error) {
	_, err = uuid.Parse(userID)
	if err != nil {
		return "", "", fmt.Errorf("invalid UUID format: %w", err)
	}

	accessTokenID := uuid.New().String()
	accessExpiration := time.Now().Add(5 * time.Minute) // todo
	accessClaims := &Claims{
		UserID:    userID,
		TokenID:   accessTokenID,
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	accessJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessJWT.SignedString(AccessTokenKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshTokenID := uuid.New().String()
	refreshExpiration := time.Now().Add(7 * 24 * time.Hour) // todo
	refreshClaims := &Claims{
		UserID:    userID,
		TokenID:   refreshTokenID,
		TokenType: RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshJWT.SignedString(RefreshTokenKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func ValidateAccessToken(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return AccessTokenKey, nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid or expired access token")
	}

	if claims.TokenType != AccessToken {
		return "", fmt.Errorf("invalid token type: expected access token")
	}

	_, err = uuid.Parse(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("token contains invalid UUID: %w", err)
	}

	return claims.UserID, nil
}

func ValidateRefreshToken(tokenString string) (string, string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return RefreshTokenKey, nil
	})

	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid or expired refresh token")
	}

	if claims.TokenType != RefreshToken {
		return "", "", fmt.Errorf("invalid token type: expected refresh token")
	}

	_, err = uuid.Parse(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("token contains invalid UUID: %w", err)
	}

	return claims.UserID, claims.TokenID, nil
}

func RefreshAccessToken(refreshToken string) (string, error) {
	userID, _, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	accessTokenID := uuid.New().String()
	accessExpiration := time.Now().Add(5 * time.Minute) // todo
	accessClaims := &Claims{
		UserID:    userID,
		TokenID:   accessTokenID,
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	accessJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	return accessJWT.SignedString(AccessTokenKey)
}
