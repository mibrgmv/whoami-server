package service

import (
	"context"
	"errors"
	"strings"

	"github.com/mibrgmv/whoami-server/shared/keycloak"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (accessToken, refreshToken, tokenType string, expiresIn int, err error)
	Register(ctx context.Context, username, email, password, firstName, lastName string) (userID, createdUsername, createdEmail string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (accessToken, newRefreshToken, tokenType string, expiresIn int, err error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	keycloak *keycloak.Client
}

func NewAuthService(keycloak *keycloak.Client) AuthService {
	return &authService{
		keycloak: keycloak,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (string, string, string, int, error) {
	if username == "" || password == "" {
		return "", "", "", 0, ErrInvalidCredentials
	}

	tokens, err := s.keycloak.ExchangeCredentialsForTokens(ctx, username, password)
	if err != nil {
		return "", "", "", 0, ErrInvalidCredentials
	}

	return tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn, nil
}

func (s *authService) Register(ctx context.Context, username, email, password, firstName, lastName string) (string, string, string, error) {
	keycloakUser := keycloak.CreateUserRequest{
		Username:      username,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		Enabled:       true,
		EmailVerified: false,
		Credentials: []keycloak.UserCredential{
			{
				Type:      "password",
				Value:     password,
				Temporary: false,
			},
		},
	}

	keycloakResp, err := s.keycloak.CreateUser(ctx, keycloakUser)
	if err != nil {
		if strings.Contains(err.Error(), "User exists with same username") {
			return "", "", "", ErrUsernameExists
		}
		if strings.Contains(err.Error(), "User exists with same email") {
			return "", "", "", ErrEmailExists
		}
		return "", "", "", err
	}

	return keycloakResp.ID, keycloakResp.Username, keycloakResp.Email, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, string, int, error) {
	tokens, err := s.keycloak.RefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", "", 0, ErrInvalidToken
	}

	return tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.keycloak.RevokeToken(ctx, refreshToken)
}
