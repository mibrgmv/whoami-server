package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/quiz"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Add(ctx context.Context, quizzes []models.Quiz) ([]models.Quiz, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				fmt.Printf("transaction rollback failed: %v\n", rbErr)
			}
			return
		}
		if cErr := tx.Commit(ctx); cErr != nil {
			fmt.Printf("transaction commit failed: %v\n", cErr)
		}
	}()

	var createdQuizzes []models.Quiz

	for _, q := range quizzes {
		query := `
		insert into quizzes (quiz_title, quiz_results)
		values ($1, $2)
		returning quiz_id`

		var createdID int64
		err = tx.QueryRow(ctx, query, q.Title, q.Results).Scan(&createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to add quiz: %w", err)
		}

		q.ID = createdID
		createdQuizzes = append(createdQuizzes, q)
	}

	return createdQuizzes, nil
}

func (r *Repository) Query(ctx context.Context, query quiz.Query) ([]models.Quiz, error) {
	sql := `
	select quiz_id,
		   quiz_title,
		   quiz_results
	from quizzes
	where ($1::bigint[] is null or cardinality($1) = 0 or quiz_id = any ($1))`

	rows, err := r.pool.Query(ctx, sql, query.Ids)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	defer rows.Close()
	var quizzes []models.Quiz
	for rows.Next() {
		var q models.Quiz
		if err := rows.Scan(&q.ID, &q.Title, &q.Results); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		quizzes = append(quizzes, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return quizzes, nil
}
