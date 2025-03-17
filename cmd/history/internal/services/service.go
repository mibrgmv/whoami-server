package services

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *Service) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	modelItems := make([]models.QuizCompletionHistoryItem, len(req.Items))
	for i, item := range req.Items {
		modelItems[i] = models.QuizCompletionHistoryItem{
			ID:         item.Id,
			QuizID:     item.QuizId,
			UserID:     item.UserId,
			QuizResult: item.QuizResult,
		}
	}

	resultItems, err := s.repo.Add(ctx, modelItems)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add items: %v", err)
	}

	responseItems := make([]*pb.QuizCompletionHistoryItem, len(resultItems))
	for i, item := range resultItems {
		responseItems[i] = &pb.QuizCompletionHistoryItem{
			Id:         item.ID,
			QuizId:     item.QuizID,
			UserId:     item.UserID,
			QuizResult: item.QuizResult,
		}
	}

	return &pb.AddResponse{Items: responseItems}, nil
}

func (s *Service) Query(req *pb.QueryRequest, stream grpc.ServerStreamingServer[pb.QuizCompletionHistoryItem]) error {
	ctx := stream.Context()

	query := Query{
		IDs:     req.Ids,
		UserIDs: req.UserIds,
		QuizIDs: req.QuizIds,
	}

	items, err := s.repo.Query(ctx, query)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to query items: %v", err)
	}

	for _, item := range items {
		protoItem := &pb.QuizCompletionHistoryItem{
			Id:         item.ID,
			QuizId:     item.QuizID,
			UserId:     item.UserID,
			QuizResult: item.QuizResult,
		}

		if err := stream.Send(protoItem); err != nil {
			return status.Errorf(codes.Internal, "failed to send response: %v", err)
		}
	}

	return nil
}
