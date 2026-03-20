package service

import (
	"context"
	"errors"
	"strings"

	"gordle/iam/internal/config"
	"gordle/libs/auth"
	"gordle/libs/keycloak"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrRecaptchaRequired  = errors.New("reCAPTCHA token is required")
	ErrRecaptchaFailed    = errors.New("reCAPTCHA verification failed")
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (accessToken, refreshToken, tokenType string, expiresIn int, err error)
	Register(ctx context.Context, username, email, password, firstName, lastName, recaptchaToken string) (userID, createdUsername, createdEmail string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (accessToken, newRefreshToken, tokenType string, expiresIn int, err error)
	Logout(ctx context.Context, refreshToken string) error
	GuestAuth(ctx context.Context) (accessToken, guestID, tokenType string, expiresIn int, err error)
}

type authService struct {
	keycloak       *keycloak.Client
	guestGenerator *auth.GuestTokenGenerator
	smtpConfigured bool
	recaptcha      *recaptchaVerifier
}

func NewAuthService(keycloak *keycloak.Client, guestSecret string, smtpHost string, recaptchaCfg config.RecaptchaConfig) AuthService {
	return &authService{
		keycloak:       keycloak,
		guestGenerator: auth.NewGuestTokenGenerator(guestSecret),
		smtpConfigured: smtpHost != "",
		recaptcha:      newRecaptchaVerifier(recaptchaCfg),
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (string, string, string, int, error) {
	if username == "" || password == "" {
		return "", "", "", 0, ErrInvalidCredentials
	}

	tokens, err := s.keycloak.ExchangeCredentialsForTokens(ctx, username, password)
	if err != nil {
		if strings.Contains(err.Error(), "Account is not fully set up") {
			return "", "", "", 0, ErrEmailNotVerified
		}
		return "", "", "", 0, ErrInvalidCredentials
	}

	return tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn, nil
}

func (s *authService) Register(ctx context.Context, username, email, password, firstName, lastName, recaptchaToken string) (string, string, string, error) {
	if err := s.recaptcha.verify(ctx, recaptchaToken); err != nil {
		return "", "", "", err
	}

	keycloakUser := keycloak.CreateUserRequest{
		Username:      username,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		Enabled:       true,
		EmailVerified: !s.smtpConfigured,
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

func (s *authService) GuestAuth(ctx context.Context) (string, string, string, int, error) {
	token, guestID, err := s.guestGenerator.GenerateToken()
	if err != nil {
		return "", "", "", 0, err
	}

	expiresIn := int(auth.GuestTokenTTL.Seconds())
	return token, guestID, "Bearer", expiresIn, nil
}
