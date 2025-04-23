package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
	"whoami-server/cmd/users/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("incorrect password")
)

type Service struct {
	users Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		users: repo,
	}
}

func (s *Service) Register(ctx context.Context, username, password string) (*models.User, error) {
	users, _ := s.users.Query(ctx, Query{Username: &username, PageSize: 1})
	if len(users) != 0 {
		return nil, fmt.Errorf("user with username %s already exists", username)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("internal error")
	}

	user := models.User{
		Name:      username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}

	createdUsers, err := s.users.Add(ctx, []models.User{user})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &createdUsers[0], nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*uuid.UUID, error) {
	users, err := s.users.Query(ctx, Query{Username: &username, PageSize: 1})
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	if len(users) == 0 {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(users[0].Password), []byte(password)); err != nil {
		return nil, ErrIncorrectPassword
	}

	users[0].LastLogin = time.Now()
	_, err = s.users.Update(ctx, []models.User{users[0]})
	if err != nil {
		return nil, fmt.Errorf("failed to update user login: %w", err)
	}

	return &users[0].ID, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.users.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	users, err := s.users.Query(ctx, Query{UserIDs: []uuid.UUID{id}, PageSize: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if len(users) == 0 {
		return nil, ErrUserNotFound
	}

	return &users[0], nil
}

func (s *Service) Get(ctx context.Context, pageSize int32, pageToken string) ([]models.User, string, error) {
	users, err := s.users.Query(ctx, Query{PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get users: %w", err)
	}

	var nextPageToken string
	if pageSize > 0 && len(users) > int(pageSize) {
		users = users[:len(users)-1]
		nextPageToken = users[len(users)-1].ID.String()
	}

	return users, nextPageToken, err
}
