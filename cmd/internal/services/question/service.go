package question

import (
	"context"
	"errors"
	"fmt"
	"whoami-server/cmd/internal/models"
)

var ErrNotFound = errors.New("no questions found for quiz")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, questions []models.Question) ([]models.Question, error) {
	return s.repo.Add(ctx, questions)
}

func (s *Service) Query(ctx context.Context, query Query) ([]models.Question, error) {
	return s.repo.Query(ctx, query)
}

func (s *Service) GetByQuizID(ctx context.Context, quizID int64) (*[]models.Question, error) {
	questions, err := s.repo.Query(ctx, Query{QuizIds: []int64{quizID}})
	if err != nil {
		return nil, err
	}

	if len(questions) == 0 {
		return nil, ErrNotFound
	}

	return &questions, nil
}

func (s *Service) EvaluateAnswers(ctx context.Context, answers []models.Answer, quiz models.Quiz) ([]float32, error) {
	questions, err := s.repo.Query(ctx, Query{QuizIds: []int64{quiz.ID}})
	if err != nil {
		return nil, err
	}

	for _, answer := range answers {
		if answer.QuizID != quiz.ID {
			return nil, errors.New("answer quiz ID does not match quiz ID")
		}
	}

	numResults := len(quiz.Results)
	results := make([]float32, numResults)

	questionsMap := make(map[int64]models.Question)
	for _, q := range questions {
		if q.QuizID != quiz.ID {
			return nil, errors.New("question quiz ID does not match quiz ID")
		}
		questionsMap[q.ID] = q
	}

	for _, answer := range answers {
		question, exists := questionsMap[answer.QuestionID]
		if !exists {
			return nil, fmt.Errorf("question with ID %d not found", answer.QuestionID)
		}

		weights, exists := question.OptionsWeights[answer.Body]
		if !exists {
			return nil, fmt.Errorf("option '%s' not found for question %d", answer.Body, question.ID)
		}

		if len(weights) != numResults {
			return nil, fmt.Errorf("weights length for option '%s' does not match number of results", answer.Body)
		}

		for i, weight := range weights {
			results[i] += weight
		}
	}

	return results, nil
}
