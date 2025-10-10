package postgresql

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mibrgmv/whoami-server/services/quizzes/internal/models"
	"github.com/mibrgmv/whoami-server/services/quizzes/internal/services/quiz"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Add(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error) {
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
	insert into quizzes (quiz_id, quiz_title, quiz_results)
	values ($1, $2, $3)
	returning quiz_id
	`

	rows, err := tx.Query(ctx, sql, uuid.New(), quiz.Title, quiz.Results)
	if err != nil {
		return nil, fmt.Errorf("failed to insert quizzes: %w", err)
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		var createdID uuid.UUID
		if err := rows.Scan(&createdID); err != nil {
			return nil, fmt.Errorf("failed to scan returned quiz_id: %w", err)
		}

		quiz.ID = createdID
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return quiz, nil
}

func (r *Repository) Query(ctx context.Context, query quiz.Query) ([]*models.Quiz, error) {
	sql := `
	select quiz_id,
		   quiz_title,
		   quiz_results
	from quizzes
	where (quiz_id > $1)
	  and ($2::uuid[] is null or cardinality($2) = 0 or quiz_id = any ($2))
	order by quiz_id asc
	limit $3
	`

	var args []interface{}

	if query.PageToken != "" {
		pageToken, err := uuid.Parse(query.PageToken)
		if err != nil {
			return nil, fmt.Errorf("invalid page token: %w", err)
		}
		args = append(args, pageToken)
	} else {
		args = append(args, uuid.Nil)
	}

	args = append(args, query.Ids)

	var pageSize int32
	if query.PageSize > 0 {
		pageSize = query.PageSize + 1
	} else {
		pageSize = query.PageSize
	}
	args = append(args, pageSize)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var quizzes []*models.Quiz
	for rows.Next() {
		q := new(models.Quiz)
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
