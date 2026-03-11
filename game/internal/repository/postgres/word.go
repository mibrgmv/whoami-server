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

func (r *wordRepo) GetSolutionByOffset(ctx context.Context, language string, offset int) (*models.Word, error) {
	sql := `
	with solution_words as (
		select word_id, word, language, is_solution,
		       row_number() over (order by word_id) - 1 as idx
		from words
		where is_solution = true and language = $1
	),
	word_count as (
		select count(*) as cnt from solution_words
	)
	select word_id, word, language, is_solution
	from solution_words, word_count
	where idx = $2 % cnt
	`

	row := r.pool.QueryRow(ctx, sql, language, offset)
	w := &models.Word{}
	err := row.Scan(&w.ID, &w.Word, &w.Language, &w.IsSolution)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no solution words available for language %s", language)
		}
		return nil, fmt.Errorf("failed to get solution by offset: %w", err)
	}
	return w, nil
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
