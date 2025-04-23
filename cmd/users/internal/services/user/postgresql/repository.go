package postgresql

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
	"whoami-server/cmd/users/internal/models"
	"whoami-server/cmd/users/internal/services/user"
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

	createdUsers := make([]models.User, 0, len(users))
	for _, u := range users {
		query := `
		insert into users (user_id, user_name, user_password, user_created_at, user_last_login)
		values ($1, $2, $3, $4, $5)
		returning user_id`

		var createdID string
		err = tx.QueryRow(ctx, query, uuid.New(), u.Name, u.Password, u.CreatedAt, u.LastLogin).Scan(&createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to add user: %w", err)
		}

		u.ID, err = uuid.Parse(createdID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse UUID: %v", err)
		}
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
	where ($1::uuid[] is null or cardinality($1) = 0 or user_id = any ($1))
	  and ($2::text is null or user_name like $2::text)
	`

	rows, err := r.pool.Query(ctx, sql, query.UserIDs, query.Username)
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

func (r Repository) Update(ctx context.Context, users []models.User) ([]models.User, error) {
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

	userIDs := make([]uuid.UUID, len(users))
	userNames := make([]string, len(users))
	userPasswords := make([]string, len(users))
	userLastLogins := make([]time.Time, len(users))

	for i, u := range users {
		userIDs[i] = u.ID
		userNames[i] = u.Name
		userPasswords[i] = u.Password
		userLastLogins[i] = u.LastLogin
	}

	sql := `
	with updated AS (
		update users
		set user_name = u.name,
			user_password = u.password,
			user_last_login = u.last_login
		from (
			select * from unnest($1::uuid[], $2::text[], $3::text[], $4::timestamptz[])
			as u(id, name, password, last_login)
		) u
		where users.user_id = u.id
		returning users.user_id, users.user_name, users.user_password, users.user_created_at, users.user_last_login
	)
	select user_id, user_name, user_password, user_created_at, user_last_login
	from updated
	`

	rows, err := tx.Query(ctx, sql, userIDs, userNames, userPasswords, userLastLogins)
	if err != nil {
		return nil, fmt.Errorf("failed to update users: %w", err)
	}
	defer rows.Close()

	updatedUsers := make([]models.User, 0, len(users))
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Password, &u.CreatedAt, &u.LastLogin); err != nil {
			return nil, fmt.Errorf("failed to scan updated user: %w", err)
		}
		updatedUsers = append(updatedUsers, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return updatedUsers, nil
}

func (r Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "delete from users where user_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
