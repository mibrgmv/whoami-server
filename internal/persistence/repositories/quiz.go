package repositories

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"whoami-server/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type QuizRepository struct {
	pool *pgxpool.Pool
}

func NewQuizRepository(pool *pgxpool.Pool) *QuizRepository {
	return &QuizRepository{pool: pool}
}

func (r *QuizRepository) AddQuizzes(ctx context.Context, quizzes []models.Quiz) ([]models.Quiz, error) {
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

	for _, quiz := range quizzes {
		query := `
		insert into quizzes (quiz_title, quiz_results)
		values ($1, $2)
		returning quiz_id`

		var createdID int64
		err = tx.QueryRow(ctx, query, quiz.Title, quiz.Results).Scan(&createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to add quiz: %w", err)
		}

		quiz.ID = createdID
		createdQuizzes = append(createdQuizzes, quiz)
	}

	return createdQuizzes, nil
}

func (r *QuizRepository) GetQuizzes(ctx context.Context) ([]models.Quiz, error) {
	query := `
	select quiz_id,
	       quiz_title
	from quizzes`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var quizzes []models.Quiz
	for rows.Next() {
		var q models.Quiz
		if err := rows.Scan(&q.ID, &q.Title); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		quizzes = append(quizzes, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return quizzes, nil
}

func (r *QuizRepository) GetQuizByID(ctx context.Context, id int64) (models.Quiz, error) {
	query := `
	select quiz_id,
	       quiz_title
	from quizzes
	where quiz_id = $1`

	row := r.pool.QueryRow(ctx, query, id)

	var quiz models.Quiz
	if err := row.Scan(&quiz.ID, &quiz.Title); err != nil {
		if err == pgx.ErrNoRows {
			return models.Quiz{}, fmt.Errorf("quiz not found")
		}
		return models.Quiz{}, fmt.Errorf("scan failed: %w", err)
	}

	return quiz, nil
}
