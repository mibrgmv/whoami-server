package history

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"whoami-server/cmd/history/internal/models"
	pb "whoami-server/protogen/golang/history"
)

type Service struct {
	repo Repository
	pb.UnimplementedQuizCompletionHistoryServiceServer
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(stream pb.QuizCompletionHistoryService_AddServer) error {
	var itemsToCreate []models.QuizCompletionHistoryItem

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("receive error %v", err)
		}

		item := models.QuizCompletionHistoryItem{
			ID:         req.Id,
			QuizID:     req.QuizId,
			UserID:     req.UserId,
			QuizResult: req.QuizResult,
		}
		itemsToCreate = append(itemsToCreate, item)
	}

	createdItems, err := s.repo.Add(stream.Context(), itemsToCreate)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to add items: %v", err)
	}

	for _, quiz := range createdItems {
		if err := stream.Send(&pb.QuizCompletionHistoryItem{
			Id:         quiz.ID,
			QuizId:     quiz.ID,
			UserId:     quiz.UserID,
			QuizResult: quiz.QuizResult,
		}); err != nil {
			log.Fatalf("can not send %v", err)
		}
	}

	return nil
}

func (s *Service) GetAll(empty *pb.Empty, stream pb.QuizCompletionHistoryService_GetAllServer) error {
	items, err := s.repo.Query(stream.Context(), Query{})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to query quizzes: %v", err)
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		for _, item := range items {
			protoItem := &pb.QuizCompletionHistoryItem{
				Id:         item.ID,
				QuizId:     item.QuizID,
				UserId:     item.UserID,
				QuizResult: item.QuizResult,
			}

			if err := stream.Send(protoItem); err != nil {
				log.Fatalf("failed to send response: %v", err)
			}
		}
	}()

	<-done
	return nil
}

func (s *Service) GetByUserID(request *pb.GetByUserIDRequest, stream pb.QuizCompletionHistoryService_GetByUserIDServer) error {
	query := Query{UserIDs: []int64{request.UserId}}

	items, err := s.repo.Query(stream.Context(), query)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to query quizzes: %v", err)
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		for _, item := range items {
			protoItem := &pb.QuizCompletionHistoryItem{
				Id:         item.ID,
				QuizId:     item.QuizID,
				UserId:     item.UserID,
				QuizResult: item.QuizResult,
			}

			if err := stream.Send(protoItem); err != nil {
				log.Fatalf("failed to send response: %v", err)
			}
		}
	}()

	<-done
	return nil
}
