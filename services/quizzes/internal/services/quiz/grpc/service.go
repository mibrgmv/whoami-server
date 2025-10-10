package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mibrgmv/whoami-server/services/quizzes/internal/models"
	pb "github.com/mibrgmv/whoami-server/services/quizzes/internal/protogen/quiz"
	"github.com/mibrgmv/whoami-server/services/quizzes/internal/services/quiz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QuizService struct {
	service *quiz.Service
	pb.UnimplementedQuizServiceServer
}

func NewService(service *quiz.Service) *QuizService {
	return &QuizService{
		service: service,
	}
}

func (s *QuizService) CreateQuiz(ctx context.Context, request *pb.CreateQuizRequest) (*pb.Quiz, error) {
	var q = &models.Quiz{
		Title:   request.Title,
		Results: request.Results,
	}

	createdQuiz, err := s.service.Add(ctx, q)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create quiz: %v", err)
	}

	return createdQuiz.ToProto(), nil
}

func (s *QuizService) GetQuiz(ctx context.Context, request *pb.GetQuizRequest) (*pb.Quiz, error) {
	quizID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid quiz ID format: %v", err)
	}

	q, err := s.service.GetByID(ctx, quizID)
	if err != nil {
		if errors.Is(err, quiz.ErrQuizNotFound) {
			return nil, status.Errorf(codes.NotFound, "quiz not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get quiz: %v", err)
	}

	return q.ToProto(), nil
}

func (s *QuizService) BatchGetQuizzes(ctx context.Context, request *pb.BatchGetQuizzesRequest) (*pb.BatchGetQuizzesResponse, error) {
	quizzes, nextPageToken, err := s.service.Get(ctx, request.PageSize, request.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get quizzes: %v", err)
	}

	var pbQuizzes []*pb.Quiz
	for _, q := range quizzes {
		pbQuizzes = append(pbQuizzes, q.ToProto())
	}

	return &pb.BatchGetQuizzesResponse{
		Quizzes:       pbQuizzes,
		NextPageToken: nextPageToken,
	}, nil
}
