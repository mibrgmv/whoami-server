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

func (r *QuestionRepository) GetQuestionsByQuizID(ctx context.Context, quizID int64) ([]models.Question, error) {
	query := `
	select question_id,
	       quiz_id,
	       question_body,
	       array_to_json(question_options)
	from questions
	where quiz_id = $1`

	rows, err := r.pool.Query(ctx, query, quizID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		var optionsStr string
		if err := rows.Scan(&q.ID, &q.QuizID, &q.Body, &optionsStr); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		if err := json.Unmarshal([]byte(optionsStr), &q.Options); err != nil {
			return nil, fmt.Errorf("json unmarshal failed: %w", err)
		}
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return questions, nil
}
