package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
	"whoami-server/cmd/users/internal/models"
	"whoami-server/internal/cache"
	"whoami-server/internal/tools"
)

type UpdateParams struct {
	ID              uuid.UUID
	User            *models.User
	CurrentPassword string
	UpdateMask      []string
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrInvalidField      = errors.New("invalid field in update mask")
)

const (
	usersCacheKeyPattern = "users:list:*"
	usersCacheKey        = "users:list:%d:%s"
)

type Service struct {
	repo  Repository
	cache cache.Interface
}

func NewService(repo Repository, cache cache.Interface) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}

func (s *Service) Create(ctx context.Context, username, password string) (*models.User, error) {
	users, _ := s.repo.Query(ctx, Query{Username: &username, PageSize: 1})
	if len(users) != 0 {
		return nil, fmt.Errorf("user with username %s already exists", username)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("internal error")
	}

	user := &models.User{
		Name:      username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		LastLogin: time.Now(),
	}

	createdUsers, err := s.repo.Add(ctx, []*models.User{user})
	if err != nil || len(createdUsers) != 1 {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.cache.DeleteByPattern(ctx, usersCacheKeyPattern); err != nil {
		fmt.Printf("failed to invalidate users cache: %v\n", err)
	}

	return createdUsers[0], nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*uuid.UUID, error) {
	users, err := s.repo.Query(ctx, Query{Username: &username, PageSize: 1})
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
	_, err = s.repo.Update(ctx, []*models.User{users[0]})
	if err != nil {
		return nil, fmt.Errorf("failed to update user login: %w", err)
	}

	return &users[0].ID, nil
}

func (s *Service) Update(ctx context.Context, params *UpdateParams) (*models.User, error) {
	existingUser, err := s.GetByID(ctx, params.ID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	// todo make global password helper tool
	// todo maybe even uuid helper tool
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(params.CurrentPassword)); err != nil {
		return nil, ErrIncorrectPassword
	}

	userToUpdate := *existingUser

	if len(params.UpdateMask) == 0 || (len(params.UpdateMask) == 1 && params.UpdateMask[0] == "*") {
		if params.User.Name != "" {
			userToUpdate.Name = params.User.Name
		}
		if params.User.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.User.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %w", err)
			}
			userToUpdate.Password = string(hashedPassword)
		}
	} else {
		for _, path := range params.UpdateMask {
			switch path {
			case "username":
				if params.User.Name != "" {
					userToUpdate.Name = params.User.Name
				}
			case "password":
				if params.User.Password != "" {
					hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.User.Password), bcrypt.DefaultCost)
					if err != nil {
						return nil, fmt.Errorf("failed to hash password: %w", err)
					}
					userToUpdate.Password = string(hashedPassword)
				}
			default:
				return nil, ErrInvalidField
			}
		}
	}

	updatedUsers, err := s.repo.Update(ctx, []*models.User{&userToUpdate})
	if err != nil || len(updatedUsers) != 1 {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUsers[0], nil
}

func (s *Service) Delete(ctx context.Context, id *uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	users, err := s.repo.Query(ctx, Query{UserIDs: []uuid.UUID{id}, PageSize: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if len(users) == 0 {
		return nil, ErrUserNotFound
	}

	return users[0], nil
}

func (s *Service) Get(ctx context.Context, pageSize int32, pageToken string) ([]*models.User, string, error) {
	parsedToken, err := tools.ParsePageToken(pageToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse token: %w", err)
	}

	cacheKey := fmt.Sprintf(usersCacheKey, pageSize, parsedToken)

	var result struct {
		Users         []*models.User `json:"users"`
		NextPageToken string         `json:"next_page_token"`
	}

	err = s.cache.Get(ctx, cacheKey, &result)
	if err == nil {
		return result.Users, result.NextPageToken, nil
	}

	users, err := s.repo.Query(ctx, Query{PageSize: pageSize, PageToken: parsedToken})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get users: %w", err)
	}

	var nextPageToken string
	if pageSize > 0 && len(users) > int(pageSize) {
		users = users[:len(users)-1]
		lastUserID := users[len(users)-1].ID
		nextPageToken = tools.CreatePageToken(lastUserID)
	}

	result.Users = users
	result.NextPageToken = nextPageToken

	if err := s.cache.Set(ctx, cacheKey, result); err != nil {
		fmt.Printf("failed to cache users: %v\n", err)
	}

	return users, nextPageToken, nil
}
