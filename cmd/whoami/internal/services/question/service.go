package question

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"whoami-server/cmd/whoami/internal/models"
	pb "whoami-server/protogen/golang/history"
)

var ErrNotFound = errors.New("no questions found for quiz")

type Service struct {
	repo          Repository
	historyClient pb.QuizCompletionHistoryServiceClient
	conn          *grpc.ClientConn
}

func NewService(repo Repository, historyServiceAddr string) (*Service, error) {
	conn, err := grpc.NewClient(historyServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to history service: %w", err)
	}

	historyClient := pb.NewQuizCompletionHistoryServiceClient(conn)

	return &Service{
		repo:          repo,
		historyClient: historyClient,
	}, nil
}

func (s *Service) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
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

func (s *Service) EvaluateAnswers(ctx context.Context, answers []models.Answer, quiz models.Quiz) (string, error) {
	questions, err := s.repo.Query(ctx, Query{QuizIds: []int64{quiz.ID}})
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

	questionsMap := make(map[int64]models.Question)
	for _, q := range questions {
		if q.QuizID != quiz.ID {
			return "", errors.New("question quiz ID does not match quiz ID")
		}
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

	item := &pb.QuizCompletionHistoryItem{
		// todo
		UserId:     1,
		QuizId:     quiz.ID,
		QuizResult: topResult,
	}

	req := &pb.AddRequest{
		Items: []*pb.QuizCompletionHistoryItem{item},
	}

	_, err = s.historyClient.Add(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to call history service Add: %w", err)
	}

	return topResult, nil
}
