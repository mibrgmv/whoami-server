package grpc

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

		q := models.Quiz{ID: req.Id, Title: req.Title, Results: req.Results}
		quizzesToCreate = append(quizzesToCreate, q)
	}

	createdQuizzes, err := s.service.Add(stream.Context(), quizzesToCreate)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to add quizzes: %v", err)
	}

	for _, q := range createdQuizzes {
		if err := stream.Send(&pb.Quiz{
			Id:      q.ID,
			Title:   q.Title,
			Results: q.Results,
		}); err != nil {
			log.Fatalf("can not send %v", err)
		}
	}

	return nil
}

func (s *QuizService) GetStream(empty *emptypb.Empty, stream pb.QuizService_GetStreamServer) error {
	quizzes, err := s.service.GetAll()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get quizzes: %v", err)
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		for _, q := range quizzes {
			if err := stream.Send(&pb.Quiz{
				Id:      q.ID,
				Title:   q.Title,
				Results: q.Results,
			}); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
	}()

	<-done
	return nil
}

func (s *QuizService) GetBatch(ctx context.Context, request *pb.GetBatchRequest) (*pb.GetBatchResponse, error) {
	quizzes, err := s.service.GetAll()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get quizzes: %v", err)
	}

	var pbQuizzes []*pb.Quiz
	for _, q := range quizzes {
		pbQuizzes = append(pbQuizzes, q.ToProto())
	}

	return &pb.GetBatchResponse{
		Quizzes:       pbQuizzes,
		NextPageToken: "", // todo
	}, nil
}

func (s *QuizService) GetByID(ctx context.Context, request *pb.GetByIDRequest) (*pb.Quiz, error) {
	q, err := s.service.GetByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, quiz.ErrQuizNotFound) {
			return nil, status.Errorf(codes.NotFound, "quiz not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to get quiz: %v", err)
	}

	return q.ToProto(), nil
}
