package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gordle/game/internal/models"
	"gordle/game/internal/repository"
)

type wordRepo struct {
	pool *pgxpool.Pool
}

func NewWordRepository(pool *pgxpool.Pool) repository.WordRepository {
	return &wordRepo{pool: pool}
}

func (r *wordRepo) GetByWord(ctx context.Context, word, language string) (*models.Word, error) {
	sql := `
	select w.word_id,
	       w.word,
	       w.language, 
	       w.is_solution
	from words w
	where w.word = $1
	  and w.language = $2
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
	select w.word_id,
	       w.word, 
	       w.language,
	       w.is_solution
	from words w
	where w.language = $1
	  and w.is_solution = true
	order by random()
	limit 1
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
	select dw.daily_word_id,
	       dw.word_id,
	       dw.language,
	       dw.game_date,
	       w.word
	from daily_words dw
	join words w
	    on dw.word_id = w.word_id
	where dw.game_date = $1
	  and dw.language = $2
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
	select exists(select 1 from words w where w.word = $1 and w.language = $2)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, sql, word, language).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check word existence: %w", err)
	}
	return exists, nil
}

func (r *wordRepo) CreateDailyWord(ctx context.Context, date, language string) (*models.DailyWord, error) {
	sql := `
	with random_word as (
		select word_id, word, $2 as language
		from words
		where is_solution = true and language = $2
		order by random()
		limit 1
	),
	inserted as (
		insert into daily_words (daily_word_id, word_id, language, game_date)
		select gen_random_uuid(), word_id, language, $1
		from random_word
		on conflict (language, game_date) do nothing
		returning daily_word_id, word_id, language, game_date
	)
	select daily_word_id, word_id, language, game_date,
	       (select word from random_word)
	from inserted
	`

	row := r.pool.QueryRow(ctx, sql, date, language)
	dw := &models.DailyWord{}
	err := row.Scan(&dw.ID, &dw.WordID, &dw.Language, &dw.GameDate, &dw.Word)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Insert returned nothing due to conflict, fetch existing
			return r.GetDailyWord(ctx, date, language)
		}
		return nil, fmt.Errorf("failed to create daily word: %w", err)
	}
	return dw, nil
}
