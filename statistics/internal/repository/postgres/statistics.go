package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"whoami-server/statistics/internal/models"
	"whoami-server/statistics/internal/repository"
)

type statisticsRepo struct {
	pool *pgxpool.Pool
}

func NewStatisticsRepository(pool *pgxpool.Pool) repository.StatisticsRepository {
	return &statisticsRepo{pool: pool}
}

func (r *statisticsRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserStatistics, error) {
	sql := `
	SELECT user_id, games_played, games_won, current_streak, max_streak,
		   guess_distribution, last_played_date, last_won_date
	FROM user_statistics
	WHERE user_id = $1
	`

	row := r.pool.QueryRow(ctx, sql, userID)
	stats := &models.UserStatistics{}

	var guessDistJSON []byte
	err := row.Scan(
		&stats.UserID,
		&stats.GamesPlayed,
		&stats.GamesWon,
		&stats.CurrentStreak,
		&stats.MaxStreak,
		&guessDistJSON,
		&stats.LastPlayedDate,
		&stats.LastWonDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &models.UserStatistics{
				UserID:            userID,
				GamesPlayed:       0,
				GamesWon:          0,
				CurrentStreak:     0,
				MaxStreak:         0,
				GuessDistribution: models.GuessDistribution{},
			}, nil
		}
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	if err := parseGuessDistribution(guessDistJSON, &stats.GuessDistribution); err != nil {
		return nil, fmt.Errorf("failed to parse guess distribution: %w", err)
	}

	return stats, nil
}

func (r *statisticsRepo) Upsert(ctx context.Context, stats *models.UserStatistics) error {
	sql := `
	INSERT INTO user_statistics (user_id, games_played, games_won, current_streak, max_streak, guess_distribution, last_played_date, last_won_date)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (user_id) DO UPDATE SET
		games_played = EXCLUDED.games_played,
		games_won = EXCLUDED.games_won,
		current_streak = EXCLUDED.current_streak,
		max_streak = EXCLUDED.max_streak,
		guess_distribution = EXCLUDED.guess_distribution,
		last_played_date = EXCLUDED.last_played_date,
		last_won_date = EXCLUDED.last_won_date
	`

	guessDistJSON := formatGuessDistribution(&stats.GuessDistribution)

	_, err := r.pool.Exec(ctx, sql,
		stats.UserID,
		stats.GamesPlayed,
		stats.GamesWon,
		stats.CurrentStreak,
		stats.MaxStreak,
		guessDistJSON,
		stats.LastPlayedDate,
		stats.LastWonDate,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert statistics: %w", err)
	}

	return nil
}

func (r *statisticsRepo) GetDailyLeaderboard(ctx context.Context, date, language string, limit int) ([]*models.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	sql := `
	SELECT gh.user_id, gh.attempts_used, gh.result
	FROM game_history gh
	WHERE gh.game_date = $1
	  AND gh.game_mode = 'daily'
	ORDER BY
		CASE WHEN gh.result = 'won' THEN 0 ELSE 1 END,
		gh.attempts_used ASC,
		gh.created_at ASC
	LIMIT $2
	`

	rows, err := r.pool.Query(ctx, sql, date, limit)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var entries []*models.LeaderboardEntry
	rank := 1
	for rows.Next() {
		e := &models.LeaderboardEntry{Rank: rank}
		var result string
		if err := rows.Scan(&e.UserID, &e.AttemptsUsed, &result); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		e.IsWon = result == "won"
		entries = append(entries, e)
		rank++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return entries, nil
}

func parseGuessDistribution(data []byte, dist *models.GuessDistribution) error {
	if len(data) == 0 {
		return nil
	}

	var raw map[string]int
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	dist.One = raw["1"]
	dist.Two = raw["2"]
	dist.Three = raw["3"]
	dist.Four = raw["4"]
	dist.Five = raw["5"]
	dist.Six = raw["6"]

	return nil
}

func formatGuessDistribution(dist *models.GuessDistribution) []byte {
	data := map[string]int{
		"1": dist.One,
		"2": dist.Two,
		"3": dist.Three,
		"4": dist.Four,
		"5": dist.Five,
		"6": dist.Six,
	}
	result, _ := json.Marshal(data)
	return result
}
