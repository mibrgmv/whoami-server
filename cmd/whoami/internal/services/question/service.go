package question

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/internal/cache"
)

const (
	questionsCacheKeyPrefix = "questions:quiz:"
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

func (s *Service) Add(ctx context.Context, quizID uuid.UUID, questions []*models.Question) ([]*models.Question, error) {
	for range questions {
		cacheKey := fmt.Sprintf("%s%s", questionsCacheKeyPrefix, quizID.String())

		err := s.cache.Delete(ctx, cacheKey)
		if err != nil {
			return nil, err
		}
	}

	return s.repo.Add(ctx, quizID, questions)
}

func (s *Service) GetByQuizID(ctx context.Context, quizID uuid.UUID) ([]models.Question, error) {
	cacheKey := fmt.Sprintf("%s%s", questionsCacheKeyPrefix, quizID.String())

	var questions []models.Question
	err := s.cache.Get(ctx, cacheKey, &questions)
	if err == nil {
		return questions, nil
	}

	questions, err = s.repo.Query(ctx, Query{QuizIds: []uuid.UUID{quizID}})
	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, cacheKey, &questions); err != nil {
		// todo ignored error
	}

	return questions, nil
}

func (s *Service) EvaluateAnswers(ctx context.Context, answers []models.Answer, quiz *models.Quiz) (string, error) {
	questions, err := s.GetByQuizID(ctx, quiz.ID)
	if err != nil {
		return "", err
	}

	for _, answer := range answers {
		if answer.QuizID != quiz.ID {
			return "", errors.New("answer quiz ID does not match quiz ID")
		}
	}

	numResults := len(quiz.Results)
	results := make([]float32, numResults)

	questionsMap := make(map[uuid.UUID]models.Question)
	for _, q := range questions {
		questionsMap[q.ID] = q
	}

	for _, answer := range answers {
		question, exists := questionsMap[answer.QuestionID]
		if !exists {
			return "", fmt.Errorf("question with ID %d not found", answer.QuestionID)
		}

		weights, exists := question.OptionsWeights[answer.Body]
		if !exists {
			return "", fmt.Errorf("option '%s' not found for question %d", answer.Body, question.ID)
		}

		if len(weights) != numResults {
			return "", fmt.Errorf("weights length for option '%s' does not match number of results", answer.Body)
		}

		for i, weight := range weights {
			results[i] += weight
		}
	}

	maxIndex := 0
	maxWeight := results[0]
	for i, weight := range results {
		if weight > maxWeight {
			maxWeight = weight
			maxIndex = i
		}
	}

	if maxIndex >= len(quiz.Results) {
		return "", fmt.Errorf("cannot map actual results to expected")
	}

	topResult := quiz.Results[maxIndex]
	return topResult, nil
}
