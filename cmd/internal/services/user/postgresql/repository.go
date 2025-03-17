package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"whoami-server/cmd/internal/models"
	"whoami-server/cmd/internal/services/user"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r Repository) Add(ctx context.Context, users []models.User) ([]models.User, error) {
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

	var createdUsers []models.User

	for _, u := range users {
		query := `
		insert into users (user_name, user_password, user_created_at, user_last_login)
		values ($1, $2, $3, $4)
		returning user_id`

		var createdID int64
		err = tx.QueryRow(ctx, query, u.Name, u.Password, u.CreatedAt, u.LastLogin).Scan(&createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to add user: %w", err)
		}

		u.ID = createdID
		createdUsers = append(createdUsers, u)
	}

	return createdUsers, nil
}

func (r Repository) Query(ctx context.Context, query user.Query) ([]models.User, error) {
	sql := `
	select user_id,
		   user_name,
		   user_password,
		   user_created_at,
		   user_last_login
	from users
	where ($1::bigint[] is null or cardinality($1) = 0 or user_id = any ($1))
	  and ($2::text is null or user_name like $2::text)`

	rows, err := r.pool.Query(ctx, sql, query.IDs, query.Username)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Password, &u.CreatedAt, &u.LastLogin); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}
