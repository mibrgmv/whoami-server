package quiz

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"whoami-server/cmd/whoami/internal/models"
	pb "whoami-server/protogen/golang/quiz"
)

var ErrNotFound = errors.New("quiz not found")

type Service struct {
	repo Repository
	pb.UnimplementedQuizServiceServer
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(stream pb.QuizService_AddServer) error {
	var quizzesToCreate []models.Quiz

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("receive error %v", err)
		}

		quiz := models.Quiz{
			ID:      req.ID,
			Title:   req.Title,
			Results: req.Results,
		}
		quizzesToCreate = append(quizzesToCreate, quiz)
	}

	createdQuizzes, err := s.repo.Add(stream.Context(), quizzesToCreate)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to add quizzes: %v", err)
	}

	for _, quiz := range createdQuizzes {
		if err := stream.Send(&pb.Quiz{
			ID:      quiz.ID,
			Title:   quiz.Title,
			Results: quiz.Results,
		}); err != nil {
			log.Fatalf("can not send %v", err)
		}
	}

	return nil
}

func (s *Service) GetAll(empty *pb.Empty, stream pb.QuizService_GetAllServer) error {
	quizzes, err := s.repo.Query(stream.Context(), Query{})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to query quizzes: %v", err)
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		for _, q := range quizzes {
			if err := stream.Send(&pb.Quiz{
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

func (s *Service) GetByID(ctx context.Context, request *pb.GetByIDRequest) (*pb.Quiz, error) {
	quizzes, err := s.repo.Query(ctx, Query{Ids: []int64{request.ID}})
	if err != nil {
		return nil, err
	}

	if len(quizzes) == 0 {
		return nil, ErrNotFound
	}

	return &pb.Quiz{
		ID:      quizzes[0].ID,
		Title:   quizzes[0].Title,
		Results: quizzes[0].Results,
	}, nil
}
