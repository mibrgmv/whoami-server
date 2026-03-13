package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gordle/game/internal/models"
	"gordle/game/internal/service"
	roomv1 "gordle/game/pkg/protogen/room/v1"
	libsgrpc "gordle/libs/grpc"
)

type roomServer struct {
	roomService service.RoomService
	roomv1.UnimplementedRoomServiceServer
}

func NewRoomServer(roomService service.RoomService) roomv1.RoomServiceServer {
	return &roomServer{
		roomService: roomService,
	}
}

func (s *roomServer) CreateRoom(ctx context.Context, req *roomv1.CreateRoomRequest) (*roomv1.CreateRoomResponse, error) {
	if libsgrpc.HasRole(ctx, "guest") {
		return nil, status.Error(codes.PermissionDenied, "guests cannot create rooms")
	}

	userID, err := libsgrpc.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required to create a room")
	}

	settings := models.RoomSettingsFromProto(req.Settings)
	language := req.Language
	if language == "" {
		language = service.DefaultLanguage
	}

	room, err := s.roomService.CreateRoom(ctx, userID.String(), language, settings)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create room: %v", err)
	}

	return &roomv1.CreateRoomResponse{Room: room.ToProto()}, nil
}

func (s *roomServer) GetRoom(ctx context.Context, req *roomv1.GetRoomRequest) (*roomv1.GetRoomResponse, error) {
	room, players, err := s.roomService.GetRoom(ctx, req.Code)
	if err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			return nil, status.Error(codes.NotFound, "room not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get room: %v", err)
	}

	protoPlayers := make([]*roomv1.RoomPlayer, len(players))
	for i, p := range players {
		protoPlayers[i] = p.ToProto()
	}

	return &roomv1.GetRoomResponse{
		Room:    room.ToProto(),
		Players: protoPlayers,
	}, nil
}

func (s *roomServer) JoinRoom(ctx context.Context, req *roomv1.JoinRoomRequest) (*roomv1.JoinRoomResponse, error) {
	userID, err := libsgrpc.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required to join a room")
	}

	if req.DisplayName == "" {
		return nil, status.Error(codes.InvalidArgument, "display_name is required")
	}

	isGuest := libsgrpc.HasRole(ctx, "guest")
	room, player, err := s.roomService.JoinRoom(ctx, req.Code, userID, isGuest, req.DisplayName)
	if err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			return nil, status.Error(codes.NotFound, "room not found")
		}
		if errors.Is(err, service.ErrRoomFull) {
			return nil, status.Error(codes.FailedPrecondition, "room is full")
		}
		if errors.Is(err, service.ErrRoomNotWaiting) {
			return nil, status.Error(codes.FailedPrecondition, "room is not accepting new players")
		}
		if errors.Is(err, service.ErrPlayerAlreadyInRoom) {
			return nil, status.Error(codes.AlreadyExists, "player is already in this room")
		}
		return nil, status.Errorf(codes.Internal, "failed to join room: %v", err)
	}

	return &roomv1.JoinRoomResponse{
		Room:   room.ToProto(),
		Player: player.ToProto(),
	}, nil
}
