package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"whoami-server/cmd/history/internal/models"
	"whoami-server/cmd/history/internal/services/history"
	pb "whoami-server/protogen/golang/history"
)

type Service struct {
	repo history.Repository
	pb.UnimplementedQuizCompletionHistoryServiceServer
}

func NewService(repo history.Repository) *Service {
	return &Service{repo: repo}
}

func (s Service) CreateItem(ctx context.Context, request *pb.CreateItemRequest) (*pb.QuizCompletionHistoryItem, error) {
	itemToCreate, err := models.ToModel(request.Item)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create history item: %v", err)
	}

	createdItems, err := s.repo.Add(ctx, []*models.QuizCompletionHistoryItem{itemToCreate})
	if err != nil {
		return nil, err
	}

	return createdItems[0].ToProto(), nil
}

func (s Service) BatchGetItems(ctx context.Context, request *pb.BatchGetItemsRequest) (*pb.BatchGetItemsResponse, error) {
	items, err := s.repo.Query(ctx, history.Query{PageSize: request.PageSize, PageToken: request.PageToken})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get history: %v", err)
	}

	var nextPageToken string
	if request.PageSize > 0 && len(items) > int(request.PageSize) {
		items = items[:len(items)-1]
		nextPageToken = items[len(items)-1].ID.String()
	}

	protoItems := make([]*pb.QuizCompletionHistoryItem, len(items))
	for i, item := range items {
		protoItems[i] = item.ToProto()
	}

	return &pb.BatchGetItemsResponse{
		Items:         protoItems,
		NextPageToken: nextPageToken,
	}, nil
}
