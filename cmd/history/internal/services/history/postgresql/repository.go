package postgresql

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"whoami-server/cmd/history/internal/models"
	"whoami-server/cmd/history/internal/services/history"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r Repository) Add(ctx context.Context, historyItems []*models.QuizCompletionHistoryItem) ([]*models.QuizCompletionHistoryItem, error) {
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

	var createdItems []*models.QuizCompletionHistoryItem
	for _, i := range historyItems {
		query := `
		insert into quiz_completion_history (quiz_completion_history_item_id, user_id, quiz_id, quiz_result)
		values ($1, $2, $3, $4)
		returning quiz_completion_history_item_id`

		var createdID string
		err = tx.QueryRow(ctx, query, uuid.New(), i.UserID, i.QuizID, i.QuizResult).Scan(&createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to add user: %w", err)
		}

		i.ID, err = uuid.Parse(createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse UUID: %v", err)
		}
		createdItems = append(createdItems, i)
	}

	return createdItems, nil
}

func (r Repository) Query(ctx context.Context, query history.Query) ([]*models.QuizCompletionHistoryItem, error) {
	sql := `
	select quiz_completion_history_item_id,
		   user_id,
		   quiz_id,
		   quiz_result
	from quiz_completion_history
	where (quiz_completion_history_item_id > $1)
	  and ($2::uuid[] is null or cardinality($2) = 0 or user_id = any ($2))
	  and ($3::uuid[] is null or cardinality($3) = 0 or quiz_id = any ($3))
	order by quiz_completion_history asc
	limit $4
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

	args = append(args, query.UserIDs, query.QuizIDs)

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
	var items []*models.QuizCompletionHistoryItem
	for rows.Next() {
		var i models.QuizCompletionHistoryItem
		if err := rows.Scan(&i.ID, &i.UserID, &i.QuizID, &i.QuizResult); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		items = append(items, &i)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}
