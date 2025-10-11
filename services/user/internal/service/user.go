package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mibrgmv/whoami-server/shared/keycloak"
	"github.com/mibrgmv/whoami-server/user/internal/service/models"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameExists  = errors.New("username already exists")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrPasswordPolicy  = errors.New("password does not meet policy requirements")
)

type UserService interface {
	GetUser(ctx context.Context, userID string) (*models.User, error)
	BatchGetUsers(ctx context.Context, pageSize, offset int32) ([]models.User, *int32, error)
	UpdateUser(ctx context.Context, userID string, update models.UpdateUserData) (*models.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	DeleteUser(ctx context.Context, userID string) error
}

type userService struct {
	keycloak *keycloak.Client
}

func NewUserService(keycloak *keycloak.Client) UserService {
	return &userService{
		keycloak: keycloak,
	}
}

func (s *userService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	kcUser, err := s.keycloak.GetUser(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &models.User{
		ID:            kcUser.ID,
		Username:      kcUser.Username,
		Email:         kcUser.Email,
		FirstName:     kcUser.FirstName,
		LastName:      kcUser.LastName,
		Enabled:       kcUser.Enabled,
		EmailVerified: kcUser.EmailVerified,
		CreatedAt:     time.Unix(kcUser.CreatedTimestamp/1000, 0).UTC().Format(time.RFC3339),
	}, nil
}

func (s *userService) BatchGetUsers(ctx context.Context, pageSize, offset int32) ([]models.User, *int32, error) {
	keycloakReq := keycloak.BatchGetUsersRequest{
		PageSize: pageSize,
		First:    offset,
	}

	keycloakResp, err := s.keycloak.BatchGetUsers(ctx, keycloakReq)
	if err != nil {
		return nil, nil, err
	}

	users := make([]models.User, len(keycloakResp.Users))
	for i, kcUser := range keycloakResp.Users {
		users[i] = models.User{
			ID:        kcUser.ID,
			Username:  kcUser.Username,
			Email:     kcUser.Email,
			FirstName: kcUser.FirstName,
			LastName:  kcUser.LastName,
			CreatedAt: time.Unix(kcUser.CreatedTimestamp/1000, 0).UTC().Format(time.RFC3339),
		}
	}

	return users, keycloakResp.NextFirst, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID string, update models.UpdateUserData) (*models.User, error) {
	updateReq := keycloak.UpdateUserRequest{
		ID: userID,
	}

	if update.Username != "" {
		updateReq.Username = update.Username
	}
	if update.Email != "" {
		updateReq.Email = update.Email
	}
	if update.FirstName != "" {
		updateReq.FirstName = update.FirstName
	}
	if update.LastName != "" {
		updateReq.LastName = update.LastName
	}

	_, err := s.keycloak.UpdateUser(ctx, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			return nil, ErrUserNotFound
		}
		if strings.Contains(err.Error(), "User exists with same username") {
			return nil, ErrUsernameExists
		}
		if strings.Contains(err.Error(), "User exists with same email") {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return s.GetUser(ctx, userID)
}

func (s *userService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if err := s.keycloak.VerifyUserPassword(ctx, userID, currentPassword); err != nil {
		return ErrInvalidPassword
	}

	if err := s.keycloak.UpdateUserPassword(ctx, userID, newPassword); err != nil {
		if strings.Contains(err.Error(), "user not found") {
			return ErrUserNotFound
		}
		if strings.Contains(err.Error(), "password policy") {
			return ErrPasswordPolicy
		}
		return err
	}

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, userID string) error {
	_, err := s.keycloak.DeleteUser(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}
