package user

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
	"whoami-server/cmd/whoami/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

type Service struct {
	users Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		users: repo,
	}
}

func (s *Service) Register(ctx context.Context, username, password string) (*models.User, error) {
	users, _ := s.users.Query(ctx, Query{Username: &username})
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

func (s *Service) Login(ctx context.Context, username, password string) (int64, error) {
	users, _ := s.users.Query(ctx, Query{Username: &username})
	if len(users) == 0 {
		return 0, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(users[0].Password), []byte(password)); err != nil {
		return 0, fmt.Errorf("invalid password")
	}

	// todo make last login update here (Upsert)

	return users[0].ID, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*models.User, error) {
	users, err := s.users.Query(ctx, Query{IDs: &[]int64{id}})
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if len(users) == 0 {
		return nil, ErrUserNotFound
	}

	return &users[0], nil
}

func (s *Service) GetAll(ctx context.Context) ([]models.User, error) {
	users, err := s.users.Query(ctx, Query{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}
