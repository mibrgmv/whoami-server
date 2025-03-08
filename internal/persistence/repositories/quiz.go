package repositories

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strconv"
	"whoami-server/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QuizRepository struct {
	pool *pgxpool.Pool
}

func NewQuizRepository(pool *pgxpool.Pool) *QuizRepository {
	return &QuizRepository{pool: pool}
}

func (r *QuizRepository) GetQuizzes(ctx context.Context) ([]models.Quiz, error) {
	query := `SELECT id, title FROM quizzes`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var quizzes []models.Quiz
	for rows.Next() {
		var q models.Quiz
		if err := rows.Scan(&q.ID, &q.Title); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		quizzes = append(quizzes, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return quizzes, nil
}

func (r *QuizRepository) GetQuizByID(ctx context.Context, id int64) (models.Quiz, error) {
	query := `SELECT id, title FROM quizzes WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)

	var quiz models.Quiz
	if err := row.Scan(&quiz.ID, &quiz.Title); err != nil {
		if err == pgx.ErrNoRows {
			return models.Quiz{}, fmt.Errorf("quiz not found")
		}
		return models.Quiz{}, fmt.Errorf("scan failed: %w", err)
	}

	return quiz, nil
}

func GetQuizzesHandler(repo *QuizRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		quizzes, err := repo.GetQuizzes(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch quizzes"})
			return
		}

		c.JSON(http.StatusOK, quizzes)
	}
}

func GetQuizByIDHandler(repo *QuizRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
			return
		}

		quiz, err := repo.GetQuizByID(c.Request.Context(), id)
		if err != nil {
			if err.Error() == "quiz not found" {
				c.JSON(http.StatusNotFound, gin.H{"message": "Quiz not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch quiz"})
			return
		}

		c.JSON(http.StatusOK, quiz)
	}
}

// Example usage (assuming you have a database connection):
//
// func main() {
//      connStr := "postgres://user:password@host:port/dbname"
//      poolConfig, err := pgxpool.ParseConfig(connStr)
//        if err != nil {
//        log.Fatalf("Unable to parse connStr: %v", err)
//        }
//      pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
//      if err != nil {
//              log.Fatal(err)
//      }
//      defer pool.Close()
//
//      repo := NewQuizRepository(pool)
//
//      r := gin.Default()
//      r.GET("/quizzes", GetQuizzesHandler(repo))
//      r.GET("/quizzes/:id", GetQuizByIDHandler(repo))
//      r.Run(":8080")
// }
