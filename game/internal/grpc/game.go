package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gordle/game/internal/models"
	"gordle/game/internal/service"
	gamev1 "gordle/game/pkg/protogen/game/v1"
	libsgrpc "gordle/libs/grpc"
)

type gameServer struct {
	gameService service.GameService
	wordService service.WordService
	gamev1.UnimplementedGameServiceServer
}

func NewGameServer(gameService service.GameService, wordService service.WordService) gamev1.GameServiceServer {
	return &gameServer{
		gameService: gameService,
		wordService: wordService,
	}
}

func (s *gameServer) StartGame(ctx context.Context, req *gamev1.StartGameRequest) (*gamev1.GameSession, error) {
	var userID *uuid.UUID
	if id, err := libsgrpc.GetUserIDFromContext(ctx); err == nil {
		userID = &id
	}

	mode := models.GameModeFromProto(req.Mode)
	language := req.Language
	if language == "" {
		language = service.DefaultLanguage
	}

	session, err := s.gameService.StartGame(ctx, userID, mode, language)
	if err != nil {
		if errors.Is(err, service.ErrGameAlreadyExists) {
			return session.ToProto(), nil
		}
		if errors.Is(err, service.ErrGuestDailyNotAllowed) {
			return nil, status.Error(codes.PermissionDenied, "guests cannot play daily mode, please register")
		}
		if errors.Is(err, service.ErrNoWordsAvailable) {
			return nil, status.Errorf(codes.FailedPrecondition, "no words available: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to start game: %v", err)
	}

	return session.ToProto(), nil
}

func (s *gameServer) SubmitGuess(ctx context.Context, req *gamev1.SubmitGuessRequest) (*gamev1.GuessResult, error) {
	var userID *uuid.UUID
	if id, err := libsgrpc.GetUserIDFromContext(ctx); err == nil {
		userID = &id
	}

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session ID: %v", err)
	}

	guess, session, err := s.gameService.SubmitGuess(ctx, userID, sessionID, req.Word)
	if err != nil {
		if errors.Is(err, service.ErrGameNotFound) {
			return nil, status.Error(codes.NotFound, "game not found")
		}
		if errors.Is(err, service.ErrNotYourGame) {
			return nil, status.Error(codes.PermissionDenied, "this game belongs to another user")
		}
		if errors.Is(err, service.ErrGameCompleted) {
			return nil, status.Error(codes.FailedPrecondition, "game is already completed")
		}
		if errors.Is(err, service.ErrMaxAttemptsReached) {
			return nil, status.Error(codes.FailedPrecondition, "maximum attempts reached")
		}
		if errors.Is(err, service.ErrInvalidWord) {
			return nil, status.Error(codes.InvalidArgument, "word is not in dictionary")
		}
		if errors.Is(err, service.ErrInvalidWordLength) {
			return nil, status.Error(codes.InvalidArgument, "word must be 5 characters")
		}
		return nil, status.Errorf(codes.Internal, "failed to submit guess: %v", err)
	}

	result := &gamev1.GuessResult{
		Guess:      guess.ToProto(),
		GameStatus: session.Status.ToProto(),
	}

	if session.Status == models.GameStatusWon || session.Status == models.GameStatusLost {
		result.TargetWord = session.TargetWord
	}

	return result, nil
}

func (s *gameServer) GetGame(ctx context.Context, req *gamev1.GetGameRequest) (*gamev1.GameSession, error) {
	var userID *uuid.UUID
	if id, err := libsgrpc.GetUserIDFromContext(ctx); err == nil {
		userID = &id
	}

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid session ID: %v", err)
	}

	session, err := s.gameService.GetGame(ctx, userID, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrGameNotFound) {
			return nil, status.Error(codes.NotFound, "game not found")
		}
		if errors.Is(err, service.ErrNotYourGame) {
			return nil, status.Error(codes.PermissionDenied, "this game belongs to another user")
		}
		return nil, status.Errorf(codes.Internal, "failed to get game: %v", err)
	}

	return session.ToProto(), nil
}

func (s *gameServer) GetDailyStatus(ctx context.Context, req *gamev1.GetDailyStatusRequest) (*gamev1.DailyStatus, error) {
	userID, err := libsgrpc.GetUserIDFromContext(ctx)
	if err != nil {
		return &gamev1.DailyStatus{HasPlayedToday: false}, nil
	}

	language := req.Language
	if language == "" {
		language = service.DefaultLanguage
	}

	session, err := s.gameService.GetDailyStatus(ctx, userID, language)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get daily status: %v", err)
	}

	result := &gamev1.DailyStatus{
		HasPlayedToday: session != nil,
	}

	if session != nil {
		result.SessionId = session.ID.String()
		result.Status = session.Status.ToProto()
	}

	return result, nil
}

func (s *gameServer) ValidateWord(ctx context.Context, req *gamev1.ValidateWordRequest) (*gamev1.ValidateWordResponse, error) {
	language := req.Language
	if language == "" {
		language = service.DefaultLanguage
	}

	valid, err := s.wordService.ValidateWord(ctx, req.Word, language)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate word: %v", err)
	}

	return &gamev1.ValidateWordResponse{IsValid: valid}, nil
}
