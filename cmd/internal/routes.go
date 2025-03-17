package internal

import (
	"github.com/gin-gonic/gin"
	questionHandler "whoami-server/cmd/internal/handlers/question"
	quizHandler "whoami-server/cmd/internal/handlers/quiz"
	"whoami-server/cmd/internal/handlers/user"
	"whoami-server/internal/jwt"
)

type RouterSetup struct {
	QuizHandler     *quizHandler.Handler
	QuestionHandler *questionHandler.Handler
	UserHandler     *user.Handler
}

func SetupRoutes(r *gin.Engine, setup RouterSetup) {
	r.POST("/login", setup.UserHandler.Login)
	r.POST("/register", setup.UserHandler.Register)

	userGroup := r.Group("u")
	{
		userGroup.Use(jwt.AuthMiddleware())
		userGroup.GET("me", setup.UserHandler.GetCurrent)
		userGroup.GET("all", setup.UserHandler.GetAll)
	}

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
