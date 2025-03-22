package internal

import (
	"github.com/gin-gonic/gin"
	"whoami-server/cmd/gateway/internal/handlers/question"
	"whoami-server/cmd/gateway/internal/handlers/quiz"
	"whoami-server/cmd/gateway/internal/handlers/user"
	"whoami-server/internal/jwt"
)

type RouterSetup struct {
	QuizHandler     *quiz.Handler
	QuestionHandler *question.Handler
	UserHandler     *user.Handler
}

func SetupRoutes(r *gin.Engine, setup RouterSetup) {
	r.POST("/login", setup.UserHandler.Login)
	r.POST("/register", setup.UserHandler.Register)

	userGroup := r.Group("/users")
	{
		userGroup.Use(jwt.AuthMiddleware())
		userGroup.GET("/current", setup.UserHandler.GetCurrent)
		userGroup.GET("", setup.UserHandler.GetAll)
	}

	quizGroup := r.Group("/quizzes")
	{
		quizGroup.GET("", setup.QuizHandler.All)
		quizGroup.POST("/add", setup.QuizHandler.Add)
		quizGroup.GET("/:id", setup.QuizHandler.GetByID)
		quizGroup.GET("/:id/questions", setup.QuestionHandler.GetByQuizID)
		quizGroup.POST("/:id/evaluate", setup.QuestionHandler.EvaluateAnswers)
	}

	questionGroup := r.Group("/questions")
	{
		questionGroup.GET("", setup.QuestionHandler.All)
		questionGroup.POST("/add", setup.QuestionHandler.Add)
	}
}
