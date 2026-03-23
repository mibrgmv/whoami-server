package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"gordle/game/internal/models"
	"gordle/game/internal/repository"
	gamev1 "gordle/game/pkg/protogen/game/v1"
	"gordle/libs/kafka"
)

var (
	ErrGameNotFound         = errors.New("game not found")
	ErrGameAlreadyExists    = errors.New("daily game already exists for today")
	ErrGameCompleted        = errors.New("game is already completed")
	ErrInvalidWord          = errors.New("word is not valid")
	ErrInvalidWordLength    = errors.New("word must be 5 characters")
	ErrNotYourGame          = errors.New("this game belongs to another user")
	ErrNoWordsAvailable     = errors.New("no words available")
	ErrMaxAttemptsReached   = errors.New("maximum attempts reached")
	ErrGuestDailyNotAllowed = errors.New("guests cannot play daily mode")
)

const (
	DefaultLanguage = "en"
	WordLength      = 5
)

type GameService interface {
	StartGame(ctx context.Context, userID *uuid.UUID, mode models.GameMode, language string) (*models.GameSession, error)
	SubmitGuess(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, word string) (*models.Guess, *models.GameSession, error)
	GetGame(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID) (*models.GameSession, error)
	GetDailyStatus(ctx context.Context, userID uuid.UUID, language string) (*models.GameSession, error)
}

type gameService struct {
	sessionRepo repository.SessionRepository
	wordService WordService
	producer    kafka.Producer
	topic       string
}

func NewGameService(
	sessionRepo repository.SessionRepository,
	wordService WordService,
	producer kafka.Producer,
	topic string,
) GameService {
	return &gameService{
		sessionRepo: sessionRepo,
		wordService: wordService,
		producer:    producer,
		topic:       topic,
	}
}

func (s *gameService) StartGame(ctx context.Context, userID *uuid.UUID, mode models.GameMode, language string) (*models.GameSession, error) {
	if language == "" {
		language = DefaultLanguage
	}

	today := time.Now().Format("2006-01-02")

	if userID == nil && mode == models.GameModeDaily {
		return nil, ErrGuestDailyNotAllowed
	}

	if mode == models.GameModeDaily && userID != nil {
		existing, err := s.sessionRepo.GetDailyByUser(ctx, *userID, language, today)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing game: %w", err)
		}
		if existing != nil {
			return existing, ErrGameAlreadyExists
		}
	}

	var targetWord string

	if mode == models.GameModeDaily {
		dailyWord, err := s.wordService.GetDailyWord(ctx, today, language)
		if err != nil {
			return nil, fmt.Errorf("failed to get daily word: %w", err)
		}
		targetWord = dailyWord.Word
	} else {
		word, err := s.wordService.GetRandomSolution(ctx, language)
		if err != nil {
			return nil, fmt.Errorf("failed to get random word: %w", err)
		}
		if word == nil {
			return nil, ErrNoWordsAvailable
		}
		targetWord = word.Word
	}

	session := &models.GameSession{
		UserID:     userID,
		Language:   language,
		GameMode:   mode,
		TargetWord: targetWord,
	}

	if mode == models.GameModeDaily {
		session.GameDate = &today
	}

	created, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	return created, nil
}

func (s *gameService) SubmitGuess(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, word string) (*models.Guess, *models.GameSession, error) {
	word = strings.ToLower(strings.TrimSpace(word))

	if len([]rune(word)) != WordLength {
		return nil, nil, ErrInvalidWordLength
	}

	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get game: %w", err)
	}
	if session == nil {
		return nil, nil, ErrGameNotFound
	}

	if session.UserID != nil && userID != nil && *session.UserID != *userID {
		return nil, nil, ErrNotYourGame
	}

	if session.Status != models.GameStatusInProgress {
		return nil, nil, ErrGameCompleted
	}

	if session.AttemptsUsed >= models.MaxAttempts {
		return nil, nil, ErrMaxAttemptsReached
	}

	valid, err := s.wordService.ValidateWord(ctx, word, session.Language)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate word: %w", err)
	}
	if !valid {
		return nil, nil, ErrInvalidWord
	}

	result := models.EvaluateGuess(word, session.TargetWord)

	guess := &models.Guess{
		ID:            uuid.New(),
		SessionID:     sessionID,
		GuessWord:     word,
		Result:        result,
		AttemptNumber: session.AttemptsUsed + 1,
	}

	session.AttemptsUsed++
	session.Guesses = append(session.Guesses, *guess)

	if word == strings.ToLower(session.TargetWord) {
		session.Status = models.GameStatusWon
		now := time.Now()
		session.CompletedAt = &now
	} else if session.AttemptsUsed >= models.MaxAttempts {
		session.Status = models.GameStatusLost
		now := time.Now()
		session.CompletedAt = &now
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("failed to update game: %w", err)
	}

	if (session.Status == models.GameStatusWon || session.Status == models.GameStatusLost) && !session.IsGuest() {
		s.publishGameCompleted(ctx, session)
	}

	return guess, session, nil
}

func (s *gameService) GetGame(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID) (*models.GameSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	if session == nil {
		return nil, ErrGameNotFound
	}

	if session.UserID != nil && userID != nil && *session.UserID != *userID {
		return nil, ErrNotYourGame
	}

	return session, nil
}

func (s *gameService) GetDailyStatus(ctx context.Context, userID uuid.UUID, language string) (*models.GameSession, error) {
	if language == "" {
		language = DefaultLanguage
	}

	today := time.Now().Format("2006-01-02")
	session, err := s.sessionRepo.GetDailyByUser(ctx, userID, language, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily game: %w", err)
	}

	return session, nil
}

func (s *gameService) publishGameCompleted(ctx context.Context, session *models.GameSession) {
	if session.UserID == nil {
		return
	}

	guessWords := make([]string, len(session.Guesses))
	for i, g := range session.Guesses {
		guessWords[i] = g.GuessWord
	}

	gameDate := ""
	if session.GameDate != nil {
		gameDate = *session.GameDate
	}

	event := &gamev1.GameCompletedEvent{
		UserId:       session.UserID.String(),
		SessionId:    session.ID.String(),
		GameMode:     string(session.GameMode),
		GameDate:     gameDate,
		TargetWord:   session.TargetWord,
		Guesses:      guessWords,
		Result:       string(session.Status),
		AttemptsUsed: int32(session.AttemptsUsed),
	}

	data, err := proto.Marshal(event)
	if err != nil {
		fmt.Printf("failed to marshal game completed event: %v\n", err)
		return
	}

	if err := s.producer.Produce(ctx, s.topic, session.ID.String(), data); err != nil {
		fmt.Printf("failed to publish game completed event: %v\n", err)
	}
}
