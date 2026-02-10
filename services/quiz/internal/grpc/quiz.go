package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mibrgmv/whoami-server/quiz/internal/models"
	quizv1 "github.com/mibrgmv/whoami-server/quiz/internal/protogen/quiz/v1"
	"github.com/mibrgmv/whoami-server/quiz/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type quizServer struct {
	quizService service.QuizService
	quizv1.UnimplementedQuizServiceServer
}

func NewQuizServer(QuizService service.QuizService) quizv1.QuizServiceServer {
	return &quizServer{
		quizService: QuizService,
	}
}

func (s *quizServer) CreateQuiz(ctx context.Context, request *quizv1.CreateQuizRequest) (*quizv1.Quiz, error) {
	var q = &models.Quiz{
		Title:   request.Title,
		Results: request.Results,
	}

	createdQuiz, err := s.quizService.Add(ctx, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create quiz: %v", err)
	}

	return createdQuiz.ToProto(), nil
}

func (s *quizServer) GetQuiz(ctx context.Context, request *quizv1.GetQuizRequest) (*quizv1.Quiz, error) {
	quizID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid quiz ID format: %v", err)
	}

	q, err := s.quizService.GetByID(ctx, quizID)
	if err != nil {
		if errors.Is(err, service.ErrQuizNotFound) {
			return nil, status.Errorf(codes.NotFound, "quiz not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get quiz: %v", err)
	}

	return q.ToProto(), nil
}

func (s *quizServer) BatchGetQuizzes(ctx context.Context, request *quizv1.BatchGetQuizzesRequest) (*quizv1.BatchGetQuizzesResponse, error) {
	quizzes, nextPageToken, err := s.quizService.Get(ctx, request.PageSize, request.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get quizzes: %v", err)
	}

	var pbQuizzes []*quizv1.Quiz
	for _, q := range quizzes {
		pbQuizzes = append(pbQuizzes, q.ToProto())
	}

	return &quizv1.BatchGetQuizzesResponse{
		Quizzes:       pbQuizzes,
		NextPageToken: nextPageToken,
	}, nil
}
