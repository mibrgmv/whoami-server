package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gordle/game/internal/models"
	"gordle/game/internal/repository"
	"gordle/libs/kafka"
)

var (
	ErrRoomNotFound         = errors.New("room not found")
	ErrRoomFull             = errors.New("room is full")
	ErrRoomNotWaiting       = errors.New("room is not in waiting state")
	ErrRoomNotPlaying       = errors.New("room is not in playing state")
	ErrNotHost              = errors.New("only the host can perform this action")
	ErrPlayerNotFound       = errors.New("player not found in room")
	ErrPlayerAlreadyInRoom  = errors.New("player is already in this room")
	ErrNotEnoughPlayers     = errors.New("not enough players to start")
	ErrPlayersNotReady      = errors.New("not all players are ready")
	ErrPlayerNotPlaying     = errors.New("player is not in playing state")
	ErrCodeGenerationFailed = errors.New("failed to generate unique room code")
)

const (
	maxCodeRetries    = 10
	minPlayersToStart = 2
)

type WSBroadcaster interface {
	BroadcastToRoom(roomCode string, event *models.WSEvent)
	SendToPlayer(roomCode, playerID string, event *models.WSEvent)
}

type RoomService interface {
	CreateRoom(ctx context.Context, hostID, language, displayName string, settings models.RoomSettings) (*models.Room, error)
	GetRoom(ctx context.Context, code string) (*models.Room, []models.RoomPlayer, error)
	JoinRoom(ctx context.Context, code string, userID uuid.UUID, isGuest bool, displayName string) (*models.Room, *models.RoomPlayer, error)
	LeaveRoom(ctx context.Context, code string, playerID string) error
	SetReady(ctx context.Context, code string, playerID string, ready bool) error
	StartGame(ctx context.Context, code string, hostID string) (*models.Room, error)
	SubmitGuess(ctx context.Context, code string, playerID string, word string) (*models.Guess, *models.RoomPlayer, *models.Room, error)
	NextRound(ctx context.Context, code string, hostID string) (*models.Room, error)
}

type roomService struct {
	roomRepo    repository.RoomRepository
	wordService WordService
	producer    kafka.Producer
	topic       string
	broadcaster WSBroadcaster

	mu         sync.Mutex
	gameTimers map[string]*time.Timer
}

func NewRoomService(
	roomRepo repository.RoomRepository,
	wordService WordService,
	producer kafka.Producer,
	topic string,
	broadcaster WSBroadcaster,
) RoomService {
	return &roomService{
		roomRepo:    roomRepo,
		wordService: wordService,
		producer:    producer,
		topic:       topic,
		broadcaster: broadcaster,
		gameTimers:  make(map[string]*time.Timer),
	}
}

func (s *roomService) CreateRoom(ctx context.Context, hostID, language, displayName string, settings models.RoomSettings) (*models.Room, error) {
	if language == "" {
		language = DefaultLanguage
	}

	room := models.NewRoom(hostID, language, settings)

	code, err := s.generateUniqueCode(ctx)
	if err != nil {
		return nil, err
	}
	room.Code = code

	created, err := s.roomRepo.Create(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	if err := s.roomRepo.SetCode(ctx, created.ID, code); err != nil {
		return nil, fmt.Errorf("failed to set room code: %w", err)
	}

	userID, err := uuid.Parse(hostID)
	if err != nil {
		return nil, fmt.Errorf("invalid host ID: %w", err)
	}
	hostPlayer := models.NewRoomPlayer(created.ID, userID, false, displayName)

	if err := s.roomRepo.AddPlayer(ctx, hostPlayer); err != nil {
		return nil, fmt.Errorf("failed to add host as player: %w", err)
	}

	return created, nil
}

func (s *roomService) generateUniqueCode(ctx context.Context) (string, error) {
	for i := 0; i < maxCodeRetries; i++ {
		code, err := models.GenerateRoomCode()
		if err != nil {
			return "", fmt.Errorf("failed to generate code: %w", err)
		}

		exists, err := s.roomRepo.CodeExists(ctx, code)
		if err != nil {
			return "", fmt.Errorf("failed to check code: %w", err)
		}

		if !exists {
			return code, nil
		}
	}
	return "", ErrCodeGenerationFailed
}

func (s *roomService) GetRoom(ctx context.Context, code string) (*models.Room, []models.RoomPlayer, error) {
	room, err := s.roomRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, nil, ErrRoomNotFound
	}

	players, err := s.roomRepo.GetPlayers(ctx, room.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get players: %w", err)
	}

	return room, players, nil
}

func (s *roomService) JoinRoom(ctx context.Context, code string, userID uuid.UUID, isGuest bool, displayName string) (*models.Room, *models.RoomPlayer, error) {
	room, err := s.roomRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, nil, ErrRoomNotFound
	}

	if room.Status != models.RoomStatusWaiting {
		return nil, nil, ErrRoomNotWaiting
	}

	playerCount, err := s.roomRepo.GetPlayerCount(ctx, room.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get player count: %w", err)
	}

	if playerCount >= room.Settings.MaxPlayers {
		return nil, nil, ErrRoomFull
	}

	player := models.NewRoomPlayer(room.ID, userID, isGuest, displayName)

	existingPlayer, err := s.roomRepo.GetPlayer(ctx, room.ID, player.PlayerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check existing player: %w", err)
	}
	if existingPlayer != nil {
		return nil, nil, ErrPlayerAlreadyInRoom
	}

	if err := s.roomRepo.AddPlayer(ctx, player); err != nil {
		return nil, nil, fmt.Errorf("failed to add player: %w", err)
	}

	s.broadcastEvent(room.Code, models.WSEventPlayerJoined, models.PlayerJoinedPayload{
		Player: player,
	})

	return room, player, nil
}

func (s *roomService) LeaveRoom(ctx context.Context, code string, playerID string) error {
	room, err := s.roomRepo.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return ErrRoomNotFound
	}

	player, err := s.roomRepo.GetPlayer(ctx, room.ID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}
	if player == nil {
		return ErrPlayerNotFound
	}

	if err := s.roomRepo.RemovePlayer(ctx, room.ID, playerID); err != nil {
		return fmt.Errorf("failed to remove player: %w", err)
	}

	s.broadcastEvent(room.Code, models.WSEventPlayerLeft, models.PlayerLeftPayload{
		PlayerID:    playerID,
		DisplayName: player.DisplayName,
	})

	if playerID == room.HostID {
		players, err := s.roomRepo.GetPlayers(ctx, room.ID)
		if err != nil {
			return fmt.Errorf("failed to get remaining players: %w", err)
		}

		if len(players) == 0 {
			if err := s.roomRepo.Delete(ctx, room.ID); err != nil {
				return fmt.Errorf("failed to delete empty room: %w", err)
			}
		} else {
			room.HostID = players[0].PlayerID
			if err := s.roomRepo.Update(ctx, room); err != nil {
				return fmt.Errorf("failed to update room host: %w", err)
			}

			s.broadcastEvent(room.Code, models.WSEventRoomUpdated, models.RoomUpdatedPayload{
				Room: room,
			})
		}
	}

	return nil
}

func (s *roomService) SetReady(ctx context.Context, code string, playerID string, ready bool) error {
	room, err := s.roomRepo.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return ErrRoomNotFound
	}

	if room.Status != models.RoomStatusWaiting {
		return ErrRoomNotWaiting
	}

	player, err := s.roomRepo.GetPlayer(ctx, room.ID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}
	if player == nil {
		return ErrPlayerNotFound
	}

	if ready {
		player.Status = models.PlayerStatusReady
	} else {
		player.Status = models.PlayerStatusWaiting
	}

	if err := s.roomRepo.UpdatePlayer(ctx, player); err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	s.broadcastEvent(room.Code, models.WSEventPlayerReady, models.PlayerReadyPayload{
		PlayerID:    playerID,
		DisplayName: player.DisplayName,
		Ready:       ready,
	})

	return nil
}

func (s *roomService) StartGame(ctx context.Context, code string, hostID string) (*models.Room, error) {
	room, err := s.roomRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	if room.HostID != hostID {
		return nil, ErrNotHost
	}

	if room.Status != models.RoomStatusWaiting {
		return nil, ErrRoomNotWaiting
	}

	players, err := s.roomRepo.GetPlayers(ctx, room.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	if len(players) < minPlayersToStart {
		return nil, ErrNotEnoughPlayers
	}

	for _, p := range players {
		if p.Status != models.PlayerStatusReady && p.PlayerID != room.HostID {
			return nil, ErrPlayersNotReady
		}
	}

	word, err := s.wordService.GetRandomSolution(ctx, room.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to get word: %w", err)
	}
	if word == nil {
		return nil, ErrNoWordsAvailable
	}

	room.CurrentWord = word.Word
	room.Status = models.RoomStatusPlaying
	room.RoundNumber = 1

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	for i := range players {
		players[i].Status = models.PlayerStatusPlaying
		players[i].CurrentAttempts = 0
		players[i].Guesses = []models.Guess{}
		players[i].Result = models.PlayerResultNone
		players[i].FinishedAt = nil
		if err := s.roomRepo.UpdatePlayer(ctx, &players[i]); err != nil {
			return nil, fmt.Errorf("failed to update player: %w", err)
		}
	}

	s.broadcastEvent(room.Code, models.WSEventGameStarted, models.GameStartedPayload{
		RoundNumber: room.RoundNumber,
		WordLength:  WordLength,
	})

	if room.Settings.Mode == models.RoomModeMarathon && room.Settings.TimeLimitSecs != nil {
		s.scheduleGameEnd(room.Code, time.Duration(*room.Settings.TimeLimitSecs)*time.Second)
	}

	return room, nil
}

func (s *roomService) SubmitGuess(ctx context.Context, code string, playerID string, word string) (*models.Guess, *models.RoomPlayer, *models.Room, error) {
	word = strings.ToLower(strings.TrimSpace(word))

	if len([]rune(word)) != WordLength {
		return nil, nil, nil, ErrInvalidWordLength
	}

	room, err := s.roomRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, nil, nil, ErrRoomNotFound
	}

	if room.Status != models.RoomStatusPlaying {
		return nil, nil, nil, ErrRoomNotPlaying
	}

	player, err := s.roomRepo.GetPlayer(ctx, room.ID, playerID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get player: %w", err)
	}
	if player == nil {
		return nil, nil, nil, ErrPlayerNotFound
	}

	if player.Status != models.PlayerStatusPlaying {
		return nil, nil, nil, ErrPlayerNotPlaying
	}

	valid, err := s.wordService.ValidateWord(ctx, word, room.Language)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to validate word: %w", err)
	}
	if !valid {
		return nil, nil, nil, ErrInvalidWord
	}

	result := models.EvaluateGuess(word, room.CurrentWord)

	guess := &models.Guess{
		ID:            uuid.New(),
		SessionID:     room.ID,
		GuessWord:     word,
		Result:        result,
		AttemptNumber: player.CurrentAttempts + 1,
	}

	player.CurrentAttempts++
	player.Guesses = append(player.Guesses, *guess)

	solved := word == strings.ToLower(room.CurrentWord)
	if solved {
		now := time.Now()
		player.Status = models.PlayerStatusFinished
		player.Result = models.PlayerResultWon
		player.FinishedAt = &now

		if room.Settings.Mode == models.RoomModeMarathon {
			player.TotalScore += s.calculateScore(player.CurrentAttempts)
		}
	} else if player.CurrentAttempts >= models.MaxAttempts {
		now := time.Now()
		player.Status = models.PlayerStatusFinished
		player.Result = models.PlayerResultLost
		player.FinishedAt = &now
	}

	if err := s.roomRepo.UpdatePlayer(ctx, player); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to update player: %w", err)
	}

	attemptPayload := models.PlayerAttemptPayload{
		PlayerID:    playerID,
		DisplayName: player.DisplayName,
		Attempts:    player.CurrentAttempts,
		Solved:      solved,
	}
	s.broadcastEvent(room.Code, models.WSEventPlayerAttempt, attemptPayload)

	if room.Settings.ShowGuesses {
		guessPayload := models.PlayerGuessPayload{
			PlayerID:    playerID,
			DisplayName: player.DisplayName,
			GuessWord:   word,
			Result:      result,
			Attempts:    player.CurrentAttempts,
			Solved:      solved,
		}
		s.broadcastEvent(room.Code, models.WSEventPlayerGuess, guessPayload)
	}

	if room.Settings.Mode == models.RoomModeSingleRound {
		if err := s.checkRoundEnd(ctx, room); err != nil {
			return nil, nil, nil, err
		}
	} else if room.Settings.Mode == models.RoomModeMarathon && solved {
		if err := s.assignNewWordForPlayer(ctx, room, player); err != nil {
			return nil, nil, nil, err
		}
	}

	updatedRoom, err := s.roomRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get updated room: %w", err)
	}

	return guess, player, updatedRoom, nil
}

func (s *roomService) calculateScore(attempts int) int {
	return models.MaxAttempts - attempts + 1
}

func (s *roomService) checkRoundEnd(ctx context.Context, room *models.Room) error {
	players, err := s.roomRepo.GetPlayers(ctx, room.ID)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	allFinished := true
	for _, p := range players {
		if p.Status == models.PlayerStatusPlaying {
			allFinished = false
			break
		}
	}

	if allFinished {
		return s.endRound(ctx, room, players)
	}

	return nil
}

func (s *roomService) endRound(ctx context.Context, room *models.Room, players []models.RoomPlayer) error {
	results := make([]models.PlayerScore, len(players))
	for i, p := range players {
		results[i] = models.PlayerScore{
			PlayerID:    p.PlayerID,
			DisplayName: p.DisplayName,
			Result:      string(p.Result),
			Attempts:    p.CurrentAttempts,
			Score:       p.TotalScore,
		}
	}

	s.broadcastEvent(room.Code, models.WSEventRoundEnded, models.RoundEndedPayload{
		RoundNumber: room.RoundNumber,
		TargetWord:  room.CurrentWord,
		Results:     results,
	})

	room.Status = models.RoomStatusFinished
	if err := s.roomRepo.Update(ctx, room); err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}

	s.broadcastEvent(room.Code, models.WSEventGameEnded, models.GameEndedPayload{
		Reason:      "round_complete",
		FinalScores: results,
		TargetWord:  room.CurrentWord,
	})

	return nil
}

func (s *roomService) assignNewWordForPlayer(ctx context.Context, room *models.Room, player *models.RoomPlayer) error {
	word, err := s.wordService.GetRandomSolution(ctx, room.Language)
	if err != nil {
		return fmt.Errorf("failed to get new word: %w", err)
	}
	if word == nil {
		return ErrNoWordsAvailable
	}

	player.Status = models.PlayerStatusPlaying
	player.CurrentAttempts = 0
	player.Guesses = []models.Guess{}
	player.Result = models.PlayerResultNone
	player.FinishedAt = nil

	if err := s.roomRepo.UpdatePlayer(ctx, player); err != nil {
		return fmt.Errorf("failed to update player for new word: %w", err)
	}

	return nil
}

func (s *roomService) NextRound(ctx context.Context, code string, hostID string) (*models.Room, error) {
	room, err := s.roomRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}

	if room.HostID != hostID {
		return nil, ErrNotHost
	}

	if room.Status != models.RoomStatusFinished {
		return nil, fmt.Errorf("room must be finished to start next round")
	}

	word, err := s.wordService.GetRandomSolution(ctx, room.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to get word: %w", err)
	}
	if word == nil {
		return nil, ErrNoWordsAvailable
	}

	room.CurrentWord = word.Word
	room.Status = models.RoomStatusPlaying
	room.RoundNumber++

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	players, err := s.roomRepo.GetPlayers(ctx, room.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	for i := range players {
		players[i].Status = models.PlayerStatusPlaying
		players[i].CurrentAttempts = 0
		players[i].Guesses = []models.Guess{}
		players[i].Result = models.PlayerResultNone
		players[i].FinishedAt = nil
		if err := s.roomRepo.UpdatePlayer(ctx, &players[i]); err != nil {
			return nil, fmt.Errorf("failed to update player: %w", err)
		}
	}

	s.broadcastEvent(room.Code, models.WSEventGameStarted, models.GameStartedPayload{
		RoundNumber: room.RoundNumber,
		WordLength:  WordLength,
	})

	return room, nil
}

func (s *roomService) scheduleGameEnd(roomCode string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.gameTimers[roomCode]; ok {
		existing.Stop()
	}

	timer := time.AfterFunc(duration, func() {
		s.handleMarathonTimeout(roomCode)
	})

	s.gameTimers[roomCode] = timer
}

func (s *roomService) handleMarathonTimeout(roomCode string) {
	ctx := context.Background()

	room, err := s.roomRepo.GetByCode(ctx, roomCode)
	if err != nil || room == nil {
		return
	}

	if room.Status != models.RoomStatusPlaying {
		return
	}

	players, err := s.roomRepo.GetPlayers(ctx, room.ID)
	if err != nil {
		return
	}

	for i := range players {
		if players[i].Status == models.PlayerStatusPlaying {
			now := time.Now()
			players[i].Status = models.PlayerStatusFinished
			players[i].FinishedAt = &now
			_ = s.roomRepo.UpdatePlayer(ctx, &players[i])
		}
	}

	room.Status = models.RoomStatusFinished
	_ = s.roomRepo.Update(ctx, room)

	results := make([]models.PlayerScore, len(players))
	for i, p := range players {
		results[i] = models.PlayerScore{
			PlayerID:    p.PlayerID,
			DisplayName: p.DisplayName,
			Result:      string(p.Result),
			Attempts:    p.CurrentAttempts,
			Score:       p.TotalScore,
		}
	}

	s.broadcastEvent(roomCode, models.WSEventGameEnded, models.GameEndedPayload{
		Reason:      "time_up",
		FinalScores: results,
	})

	s.mu.Lock()
	delete(s.gameTimers, roomCode)
	s.mu.Unlock()
}

func (s *roomService) broadcastEvent(roomCode string, eventType models.WSEventType, payload any) {
	if s.broadcaster == nil {
		return
	}

	event, err := models.NewWSEvent(eventType, payload)
	if err != nil {
		return
	}

	s.broadcaster.BroadcastToRoom(roomCode, event)
}
