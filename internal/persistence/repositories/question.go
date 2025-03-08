package repositories

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"whoami-server/internal/models"

	"github.com/gin-gonic/gin"
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
	select id,
	       quiz_id,
	       body,
	       options
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
		q.Options = parseOptionsString(optionsStr)
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return questions, nil
}

// Helper function to parse the options string (e.g., "{Yes,No}") into a slice.
func parseOptionsString(optionsStr string) []string {
	if len(optionsStr) < 3 {
		return []string{}
	}

	optionsStr = optionsStr[1 : len(optionsStr)-1]
	var options []string
	var currentOption string
	inQuotes := false

	for _, char := range optionsStr {
		if char == ',' && !inQuotes {
			options = append(options, currentOption)
			currentOption = ""
		} else if char == '"' {
			inQuotes = !inQuotes
			currentOption += string(char)
		} else {
			currentOption += string(char)
		}
	}

	if currentOption != "" {
		options = append(options, currentOption)
	}

	return options
}

func GetQuestionsByQuizIDHandler(repo *QuestionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		quizIDStr := c.Param("id")
		quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid quiz ID"})
			return
		}

		questions, err := repo.GetQuestionsByQuizID(c.Request.Context(), quizID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch questions"})
			return
		}

		if len(questions) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "No questions found for the given quiz ID"})
			return
		}

		c.JSON(http.StatusOK, questions)
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
//      repo := NewQuestionRepository(pool)
//
//      r := gin.Default()
//      r.GET("/quizzes/:id/questions", GetQuestionsByQuizIDHandler(repo))
//      r.Run(":8080")
// }
