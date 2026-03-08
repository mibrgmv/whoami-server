package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"whoami-server/statistics/internal/models"
	"whoami-server/statistics/internal/repository"
)

const defaultPageSize int32 = 50

type historyRepo struct {
	pool *pgxpool.Pool
}

func NewHistoryRepository(pool *pgxpool.Pool) repository.HistoryRepository {
	return &historyRepo{pool: pool}
}

func (r *historyRepo) Create(ctx context.Context, history *models.GameHistory) (*models.GameHistory, error) {
	sql := `
	insert into game_history (history_id, user_id, session_id, game_mode, game_date, target_word, guesses, result, attempts_used, created_at)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	returning history_id
	`

	history.ID = uuid.New()
	_, err := r.pool.Exec(ctx, sql,
		history.ID,
		history.UserID,
		history.SessionID,
		history.GameMode,
		history.GameDate,
		history.TargetWord,
		history.Guesses,
		history.Result,
		history.AttemptsUsed,
		history.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert game history: %w", err)
	}

	return history, nil
}

func (r *historyRepo) Query(ctx context.Context, query models.GameHistoryQuery) ([]*models.GameHistory, error) {
	sql := `
	select history_id,
	       user_id,
	       session_id,
	       game_mode,
	       game_date,
	       target_word,
	       guesses,
	       result,
	       attempts_used,
	       created_at
	from game_history
	where user_id = $1
	  and ($2::uuid is null or history_id < $2)
	order by created_at desc
	limit $3
	`

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	pageSize++ // Fetch one extra to check for next page

	var args []interface{}
	args = append(args, query.UserID)

	if query.PageToken != "" {
		pageToken, err := uuid.Parse(query.PageToken)
		if err != nil {
			return nil, fmt.Errorf("invalid page token: %w", err)
		}
		args = append(args, pageToken)
	} else {
		args = append(args, nil)
	}

	args = append(args, pageSize)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var history []*models.GameHistory
	for rows.Next() {
		h := &models.GameHistory{}
		if err := rows.Scan(
			&h.ID,
			&h.UserID,
			&h.SessionID,
			&h.GameMode,
			&h.GameDate,
			&h.TargetWord,
			&h.Guesses,
			&h.Result,
			&h.AttemptsUsed,
			&h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return history, nil
}
