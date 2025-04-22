package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/question"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Add(ctx context.Context, questions []models.Question) ([]models.Question, error) {
	args := m.Called(ctx, questions)
	return args.Get(0).([]models.Question), args.Error(1)
}

func (m *MockRepository) Query(ctx context.Context, query question.Query) ([]models.Question, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]models.Question), args.Error(1)
}
