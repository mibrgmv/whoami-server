package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"whoami-server/cmd/whoami/internal/models"
	"whoami-server/cmd/whoami/internal/services/question"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Add(ctx context.Context, quizID uuid.UUID, questions []*models.Question) ([]*models.Question, error) {
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

	var createdQuestions []*models.Question

	for _, q := range questions {
		optionsWeightsJSON, err := json.Marshal(q.OptionsWeights)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal options_weights: %w", err)
		}

		query := `
		insert into questions (question_id, quiz_id, question_body, question_options_weights)
		values ($1, $2, $3, $4)
		returning question_id`

		var createdID string
		err = tx.QueryRow(ctx, query, uuid.New(), quizID, q.Body, optionsWeightsJSON).Scan(&createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to add question: %w", err)
		}

		q.ID, err = uuid.Parse(createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse UUID: %v", err)
		}
		createdQuestions = append(createdQuestions, q)
	}

	return createdQuestions, nil
}

func (r *Repository) Query(ctx context.Context, query question.Query) ([]*models.Question, error) {
	sql := `
	select question_id,
		   question_body,
		   question_options_weights
	from questions
	where ($1::uuid[] is null or cardinality($1) = 0 or quiz_id = any ($1))`

	rows, err := r.pool.Query(ctx, sql, query.QuizIds)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		q := new(models.Question)
		var optionsWeightsJSON []byte

		if err := rows.Scan(&q.ID, &q.Body, &optionsWeightsJSON); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		if err = json.Unmarshal(optionsWeightsJSON, &q.OptionsWeights); err != nil {
			return nil, fmt.Errorf("unmarshal failed: %w", err)
		}

		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return questions, nil
}
