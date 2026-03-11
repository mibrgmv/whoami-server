package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gordle/game/internal/models"
	"gordle/game/internal/repository"
)

const (
	wordExistsCachePrefix = "wordle:word_exists:"
	wordExistsCacheTTL    = 24 * time.Hour
)

type WordService interface {
	ValidateWord(ctx context.Context, word, language string) (bool, error)
	GetRandomSolution(ctx context.Context, language string) (*models.Word, error)
	GetDailyWord(ctx context.Context, date, language string) (*models.Word, error)
}

type wordService struct {
	wordRepo    repository.WordRepository
	redisClient *goredis.Client
}

func NewWordService(wordRepo repository.WordRepository, redisClient *goredis.Client) WordService {
	return &wordService{
		wordRepo:    wordRepo,
		redisClient: redisClient,
	}
}

func (s *wordService) ValidateWord(ctx context.Context, word, language string) (bool, error) {
	cacheKey := fmt.Sprintf("%s%s:%s", wordExistsCachePrefix, language, word)
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		return cached == "1", nil
	}
	if err != nil && !errors.Is(err, goredis.Nil) {
		fmt.Printf("redis cache error: %v\n", err)
	}

	exists, err := s.wordRepo.WordExists(ctx, word, language)
	if err != nil {
		return false, fmt.Errorf("failed to check word: %w", err)
	}

	cacheValue := "0"
	if exists {
		cacheValue = "1"
	}
	_ = s.redisClient.Set(ctx, cacheKey, cacheValue, wordExistsCacheTTL).Err()

	return exists, nil
}

func (s *wordService) GetRandomSolution(ctx context.Context, language string) (*models.Word, error) {
	return s.wordRepo.GetRandomSolution(ctx, language)
}

func (s *wordService) GetDailyWord(ctx context.Context, date, language string) (*models.Word, error) {
	offset := hashDateLanguage(date, language)
	return s.wordRepo.GetSolutionByOffset(ctx, language, offset)
}

func hashDateLanguage(date, language string) int {
	h := fnv.New32a()
	h.Write([]byte(date + language))
	return int(h.Sum32())
}
