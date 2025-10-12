package postgresql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mibrgmv/whoami-server/quiz/internal/models"
	"github.com/mibrgmv/whoami-server/quiz/internal/service/question"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Add(ctx context.Context, questions []*models.Question) ([]*models.Question, error) {
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

	sql := `
	insert into questions (question_id, quiz_id, question_body, question_options_weights)
	select question_id,
	       quiz_id,
	       question_body,
	       question_options_weights
	from unnest($1::uuid[], $2::uuid[], $3::text[], $4::jsonb[])
	    as source (question_id, quiz_id, question_body, question_options_weights)
	returning question_id
	`

	questionIDs := make([]uuid.UUID, len(questions))
	quizIDs := make([]uuid.UUID, len(questions))
	bodies := make([]string, len(questions))
	optionWeights := make([][]byte, len(questions))

	for i, q := range questions {
		questionIDs[i] = uuid.New()
		quizIDs[i] = q.QuizID
		bodies[i] = q.Body
		optionsWeightsJSON, err := json.Marshal(q.OptionsWeights)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal options_weights: %w", err)
		}
		optionWeights[i] = optionsWeightsJSON
	}

	rows, err := tx.Query(ctx, sql, questionIDs, quizIDs, bodies, optionWeights)
	if err != nil {
		return nil, fmt.Errorf("failed to insert questions: %w", err)
	}
	defer rows.Close()

	createdQuestions := make([]*models.Question, 0, len(questions))
	for i := 0; rows.Next(); i++ {
		var createdID uuid.UUID
		if err := rows.Scan(&createdID); err != nil {
			return nil, fmt.Errorf("failed to scan returned question_id: %w", err)
		}

		questions[i].ID = createdID
		createdQuestions = append(createdQuestions, questions[i])
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return createdQuestions, nil
}

func (r *Repository) Query(ctx context.Context, query question.Query) ([]*models.Question, error) {
	sql := `
	select question_id,
	       quiz_id,
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
