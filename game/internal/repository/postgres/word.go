package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"whoami-server/game/internal/models"
	"whoami-server/game/internal/repository"
)

type wordRepo struct {
	pool *pgxpool.Pool
}

func NewWordRepository(pool *pgxpool.Pool) repository.WordRepository {
	return &wordRepo{pool: pool}
}

func (r *wordRepo) GetByWord(ctx context.Context, word, language string) (*models.Word, error) {
	sql := `
	SELECT word_id, word, language, is_solution
	FROM words
	WHERE word = $1 AND language = $2
	`

	row := r.pool.QueryRow(ctx, sql, word, language)
	w := &models.Word{}
	err := row.Scan(&w.ID, &w.Word, &w.Language, &w.IsSolution)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get word: %w", err)
	}
	return w, nil
}

func (r *wordRepo) GetRandomSolution(ctx context.Context, language string) (*models.Word, error) {
	sql := `
	SELECT word_id, word, language, is_solution
	FROM words
	WHERE language = $1 AND is_solution = true
	ORDER BY RANDOM()
	LIMIT 1
	`

	row := r.pool.QueryRow(ctx, sql, language)
	w := &models.Word{}
	err := row.Scan(&w.ID, &w.Word, &w.Language, &w.IsSolution)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no solution words available for language %s", language)
		}
		return nil, fmt.Errorf("failed to get random solution: %w", err)
	}
	return w, nil
}

func (r *wordRepo) GetDailyWord(ctx context.Context, date, language string) (*models.DailyWord, error) {
	sql := `
	SELECT dw.daily_word_id, dw.word_id, dw.language, dw.game_date, w.word
	FROM daily_words dw
	JOIN words w ON dw.word_id = w.word_id
	WHERE dw.game_date = $1 AND dw.language = $2
	`

	row := r.pool.QueryRow(ctx, sql, date, language)
	dw := &models.DailyWord{}
	err := row.Scan(&dw.ID, &dw.WordID, &dw.Language, &dw.GameDate, &dw.Word)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get daily word: %w", err)
	}
	return dw, nil
}

func (r *wordRepo) WordExists(ctx context.Context, word, language string) (bool, error) {
	sql := `
	SELECT EXISTS(SELECT 1 FROM words WHERE word = $1 AND language = $2)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, sql, word, language).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check word existence: %w", err)
	}
	return exists, nil
}
