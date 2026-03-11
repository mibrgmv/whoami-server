package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	libsgrpc "gordle/libs/grpc"
	"gordle/statistics/internal/service"
	statisticsv1 "gordle/statistics/pkg/protogen/statistics/v1"
)

type statisticsServer struct {
	historyService    service.HistoryService
	statisticsService service.StatisticsService
	statisticsv1.UnimplementedStatisticsServiceServer
}

func NewStatisticsServer(historyService service.HistoryService, statisticsService service.StatisticsService) statisticsv1.StatisticsServiceServer {
	return &statisticsServer{
		historyService:    historyService,
		statisticsService: statisticsService,
	}
}

func (s *statisticsServer) GetMyStatistics(ctx context.Context, req *statisticsv1.GetMyStatisticsRequest) (*statisticsv1.UserStatistics, error) {
	userID, err := libsgrpc.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID: %v", err)
	}

	stats, err := s.statisticsService.GetUserStatistics(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get statistics: %v", err)
	}

	return stats.ToProto(), nil
}

func (s *statisticsServer) GetUserStatistics(ctx context.Context, req *statisticsv1.GetUserStatisticsRequest) (*statisticsv1.UserStatistics, error) {
	_, err := libsgrpc.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID: %v", err)
	}

	targetUserID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	stats, err := s.statisticsService.GetUserStatistics(ctx, targetUserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get statistics: %v", err)
	}

	return stats.ToProto(), nil
}

func (s *statisticsServer) GetMyHistory(ctx context.Context, req *statisticsv1.GetMyHistoryRequest) (*statisticsv1.GetMyHistoryResponse, error) {
	userID, err := libsgrpc.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID: %v", err)
	}

	history, nextPageToken, err := s.historyService.GetUserHistory(ctx, userID, req.PageSize, req.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get history: %v", err)
	}

	var items []*statisticsv1.GameHistoryItem
	for _, h := range history {
		items = append(items, h.ToProto())
	}

	return &statisticsv1.GetMyHistoryResponse{
		Items:         items,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *statisticsServer) GetDailyLeaderboard(ctx context.Context, req *statisticsv1.GetDailyLeaderboardRequest) (*statisticsv1.GetDailyLeaderboardResponse, error) {
	_, err := libsgrpc.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID: %v", err)
	}

	date, entries, err := s.statisticsService.GetDailyLeaderboard(ctx, req.Language, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get leaderboard: %v", err)
	}

	var protoEntries []*statisticsv1.LeaderboardEntry
	for _, e := range entries {
		protoEntries = append(protoEntries, e.ToProto())
	}

	return &statisticsv1.GetDailyLeaderboardResponse{
		GameDate: date,
		Entries:  protoEntries,
	}, nil
}
