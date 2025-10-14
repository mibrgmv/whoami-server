package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/mibrgmv/whoami-server/history/internal/models"
	historyv1 "github.com/mibrgmv/whoami-server/history/internal/protogen/history/v1"
	"github.com/mibrgmv/whoami-server/history/internal/service"
	"github.com/mibrgmv/whoami-server/shared/grpc/interceptor"
	"github.com/mibrgmv/whoami-server/shared/tools"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type historyServiceServer struct {
	service service.HistoryService
	historyv1.UnimplementedHistoryServiceServer
}

func NewHistoryServiceServer(service service.HistoryService) historyv1.HistoryServiceServer {
	return &historyServiceServer{
		service: service,
	}
}

func (s *historyServiceServer) CreateItem(ctx context.Context, req *historyv1.CreateItemRequest) (*historyv1.QuizCompletionHistoryItem, error) {
	itemToCreate, err := models.ToModel(req.Item)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create history item: %v", err)
	}

	createdItems, err := s.service.CreateItem(ctx, itemToCreate)
	if err != nil {
		return nil, err
	}

	return createdItems[0].ToProto(), nil
}

func (s *historyServiceServer) BatchGetMyItems(ctx context.Context, req *historyv1.BatchGetMyItemsRequest) (*historyv1.BatchGetItemsResponse, error) {
	parsedToken, err := tools.ParsePageToken(req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse page token: %v", err)
	}

	userID, err := interceptor.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID from context: %v", err)
	}

	quizIDs, err := parseUUIDs(req.QuizIds)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse quiz IDs: %v", err)
	}

	items, nextPageToken, err := s.service.GetItems(ctx, []*uuid.UUID{&userID}, quizIDs, req.PageSize, parsedToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get history: %v", err)
	}

	protoItems := make([]*historyv1.QuizCompletionHistoryItem, len(items))
	for i, item := range items {
		protoItems[i] = item.ToProto()
	}

	return &historyv1.BatchGetItemsResponse{
		Items:         protoItems,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *historyServiceServer) BatchGetItems(ctx context.Context, req *historyv1.BatchGetItemsRequest) (*historyv1.BatchGetItemsResponse, error) {
	parsedToken, err := tools.ParsePageToken(req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse page token: %v", err)
	}

	userIDs, err := parseUUIDs(req.UserIds)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse user IDs: %v", err)
	}

	quizIDs, err := parseUUIDs(req.QuizIds)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse quiz IDs: %v", err)
	}

	items, nextPageToken, err := s.service.GetItems(ctx, userIDs, quizIDs, req.PageSize, parsedToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get history: %v", err)
	}

	protoItems := make([]*historyv1.QuizCompletionHistoryItem, len(items))
	for i, item := range items {
		protoItems[i] = item.ToProto()
	}

	return &historyv1.BatchGetItemsResponse{
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
