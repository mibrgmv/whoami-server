package grpc

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/quiz"
	pb "whoami-server/protogen/golang/quiz"
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

func (s *QuizService) AddStream(stream pb.QuizService_AddStreamServer) error {
	var quizzesToCreate []models.Quiz

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("receive error %v", err)
		}

		q, err := models.QuizToModel(req)
		if err != nil {
			log.Fatalf("parse error %v", err)
		}

		quizzesToCreate = append(quizzesToCreate, *q)
	}

	createdQuizzes, err := s.service.Add(stream.Context(), quizzesToCreate)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to add quizzes: %v", err)
	}

	for _, q := range createdQuizzes {
		if err := stream.Send(q.ToProto()); err != nil {
			log.Fatalf("can not send %v", err)
		}
	}

	return nil
}

func (s *QuizService) GetBatch(ctx context.Context, request *pb.GetBatchRequest) (*pb.GetBatchResponse, error) {
	quizzes, nextPageToken, err := s.service.Get(ctx, request.PageSize, request.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get quizzes: %v", err)
	}

	var pbQuizzes []*pb.Quiz
	for _, q := range quizzes {
		pbQuizzes = append(pbQuizzes, q.ToProto())
	}

	return &pb.GetBatchResponse{
		Quizzes:       pbQuizzes,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *QuizService) GetByID(ctx context.Context, request *pb.GetByIDRequest) (*pb.Quiz, error) {
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
