package user

import (
	"context"
	"github.com/google/uuid"
	"whoami-server/cmd/users/internal/models"
)

type Repository interface {
	Add(ctx context.Context, users []models.User) ([]models.User, error)
	Query(ctx context.Context, query Query) ([]models.User, error)
	Update(ctx context.Context, users []models.User) ([]models.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
