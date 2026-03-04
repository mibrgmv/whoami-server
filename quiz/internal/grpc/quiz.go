package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	libsgrpc "whoami-server/libs/grpc"
	"whoami-server/quiz/internal/models"
	"whoami-server/quiz/internal/service"
	quizv1 "whoami-server/quiz/pkg/protogen/quiz/v1"
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

func (s *quizServer) DeleteQuiz(ctx context.Context, request *quizv1.DeleteQuizRequest) (*quizv1.DeleteQuizResponse, error) {
	if !libsgrpc.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin role required")
	}

	quizID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid quiz ID format: %v", err)
	}

	if err := s.quizService.Delete(ctx, quizID); err != nil {
		if errors.Is(err, service.ErrQuizNotFound) {
			return nil, status.Errorf(codes.NotFound, "quiz not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete quiz: %v", err)
	}

	return &quizv1.DeleteQuizResponse{
		Id:      request.Id,
		Message: "quiz deleted successfully",
	}, nil
}
