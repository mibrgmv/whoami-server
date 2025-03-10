package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"whoami-server/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type QuestionRepository struct {
	pool *pgxpool.Pool
}

func NewQuestionRepository(pool *pgxpool.Pool) *QuestionRepository {
	return &QuestionRepository{pool: pool}
}

func (r *QuestionRepository) AddQuestions(ctx context.Context, questions []models.Question) ([]models.Question, error) {
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

	var createdQuestions []models.Question

	for _, question := range questions {
		optionsWeightsJSON, err := json.Marshal(question.OptionsWeights)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal options_weights: %w", err)
		}

		query := `
		insert into questions (quiz_id, question_body, question_options_weights)
		values ($1, $2, $3)
		returning question_id`

		var createdID int64
		err = tx.QueryRow(ctx, query, question.QuizID, question.Body, optionsWeightsJSON).Scan(&createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to add question: %w", err)
		}

		question.ID = createdID
		createdQuestions = append(createdQuestions, question)
	}

	return createdQuestions, nil
}

func (r *QuestionRepository) QueryQuestions(ctx context.Context, query QuestionQuery) ([]models.Question, error) {
	sql := `
	select question_id,
		   quiz_id,
		   question_body,
		   question_options_weights
	from questions
	where ($1::bigint[] is null or cardinality($1) = 0 or quiz_id = any ($1))`

	rows, err := r.pool.Query(ctx, sql, query.QuizIds)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		var optionsWeightsJSON []byte

		if err := rows.Scan(&q.ID, &q.QuizID, &q.Body, &optionsWeightsJSON); err != nil {
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
