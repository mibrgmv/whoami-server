package grpc

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"whoami-server/cmd/history/internal/models"
	"whoami-server/cmd/history/internal/services/history"
	"whoami-server/internal/tools"
	pb "whoami-server/protogen/golang/history"
)

type Service struct {
	service *history.Service
	pb.UnimplementedQuizCompletionHistoryServiceServer
}

func NewService(service *history.Service) *Service {
	return &Service{service: service}
}

func (s Service) CreateItem(ctx context.Context, request *pb.CreateItemRequest) (*pb.QuizCompletionHistoryItem, error) {
	itemToCreate, err := models.ToModel(request.Item)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create history item: %v", err)
	}

	createdItems, err := s.service.CreateItem(ctx, itemToCreate)
	if err != nil {
		return nil, err
	}

	return createdItems[0].ToProto(), nil
}

func (s Service) BatchGetItems(ctx context.Context, request *pb.BatchGetItemsRequest) (*pb.BatchGetItemsResponse, error) {
	parsedToken, err := tools.ParsePageToken(request.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse page token: %v", err)
	}

	userIDs, err := parseUUIDs(request.UserIds)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse user IDs: %v", err)
	}

	quizIDs, err := parseUUIDs(request.QuizIds)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse quiz IDs: %v", err)
	}

	items, nextPageToken, err := s.service.GetItems(ctx, userIDs, quizIDs, request.PageSize, parsedToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get history: %v", err)
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

func parseUUIDs(values []*wrapperspb.StringValue) ([]*uuid.UUID, error) {
	var uuids []*uuid.UUID
	for _, u := range values {
		if u != nil {
			parsedUUID, err := uuid.Parse(u.Value)
			if err != nil {
				return nil, err
			}
			uuids = append(uuids, &parsedUUID)
		}
	}

	return uuids, nil
}
