package user

import (
	"context"
	"whoami-server/cmd/whoami/internal/models"
)

type Repository interface {
	Add(ctx context.Context, users []models.User) ([]models.User, error)
	Query(ctx context.Context, query Query) ([]models.User, error)
}
