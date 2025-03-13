package internal

import (
	"github.com/gin-gonic/gin"
	questionHandler "whoami-server/cmd/internal/handlers/question"
	quizHandler "whoami-server/cmd/internal/handlers/quiz"
)

type RouterSetup struct {
	QuizHandler     *quizHandler.Handler
	QuestionHandler *questionHandler.Handler
}

func SetupRoutes(r *gin.Engine, setup RouterSetup) {
	quizGroup := r.Group("/quiz")
	{
		quizGroup.GET("/q", setup.QuizHandler.Query)
		quizGroup.POST("/a", setup.QuizHandler.Add)
		quizGroup.GET("/:id", setup.QuizHandler.GetByID)
		quizGroup.GET("/:id/questions", setup.QuestionHandler.GetByQuizID)
		quizGroup.POST("/:id/evaluate", setup.QuizHandler.EvaluateAnswers)
	}

	questionGroup := r.Group("/question")
	{
		questionGroup.GET("/q", setup.QuestionHandler.Query)
		questionGroup.POST("/a", setup.QuestionHandler.Add)
	}
}
