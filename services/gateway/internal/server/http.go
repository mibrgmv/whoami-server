package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	appcfg "github.com/mibrgmv/whoami-server/gateway/internal/config"
	"github.com/mibrgmv/whoami-server/gateway/internal/middleware"
	authv1 "github.com/mibrgmv/whoami-server/gateway/internal/protogen/auth/v1"
	historyv1 "github.com/mibrgmv/whoami-server/gateway/internal/protogen/history/v1"
	questionv1 "github.com/mibrgmv/whoami-server/gateway/internal/protogen/question/v1"
	quizv1 "github.com/mibrgmv/whoami-server/gateway/internal/protogen/quiz/v1"
	userv1 "github.com/mibrgmv/whoami-server/gateway/internal/protogen/user/v1"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func NewHttpServer(ctx context.Context, cfg appcfg.Config) (*http.Server, error) {
	gwmux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			md := metadata.New(map[string]string{
				"authorization": req.Header.Get("Authorization"),
			})

			if userId, ok := ctx.Value("user_id").(string); ok && userId != "" {
				md.Set("user_id", userId)
			}

			if username, ok := ctx.Value("username").(string); ok && username != "" {
				md.Set("username", username)
			}

			if email, ok := ctx.Value("email").(string); ok && email != "" {
				md.Set("email", email)
			}

			return md
		}),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := authv1.RegisterAuthServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.AuthService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register auth service: %w", err)
	}

	if err := quizv1.RegisterQuizServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.QuizService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register quiz service: %w", err)
	}

	if err := questionv1.RegisterQuestionServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.QuizService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register question service: %w", err)
	}

	if err := userv1.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.UserService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register user service: %w", err)
	}

	if err := historyv1.RegisterQuizCompletionHistoryServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.HistoryService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register history service: %w", err)
	}

	jwtMiddleware := middleware.JWT(middleware.JWTConfig{
		KeycloakBaseURL: cfg.Keycloak.BaseURL,
		Realm:           cfg.Keycloak.Realm,
		KeyRefreshTTL:   1 * time.Hour,
		HTTPTimeout:     10 * time.Second,
	})

	switch cfg.HTTP.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
		log.Println("Running in debug mode")
		if configJSON, err := json.MarshalIndent(cfg, "", "  "); err == nil {
			log.Printf("Configuration:\n%s", configJSON)
		}
	case "release":
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in release mode")
	default:
		gin.SetMode(gin.ReleaseMode)
		log.Printf("Unknown mode '%s', defaulting to release mode", cfg.HTTP.Mode)
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.HTTP.CORS.AllowedOrigins,
		AllowMethods:     cfg.HTTP.CORS.AllowedMethods,
		AllowHeaders:     cfg.HTTP.CORS.AllowedHeaders,
		ExposeHeaders:    cfg.HTTP.CORS.ExposeHeaders,
		AllowCredentials: cfg.HTTP.CORS.AllowCredentials,
		MaxAge:           time.Duration(cfg.HTTP.CORS.MaxAge) * time.Second,
	}))

	router.GET("/api/v1/swagger.json", func(c *gin.Context) {
		c.File("./api/v1/gateway.swagger.json")
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/api/v1/swagger.json")))

	router.Any("/api/v1/auth/*path", gin.WrapH(gwmux))

	gwmuxGroup := router.Group("/api/v1")
	gwmuxGroup.Use(jwtMiddleware)
	{
		gwmuxGroup.Any("/quizzes", gin.WrapH(gwmux))
		gwmuxGroup.Any("/quizzes/*path", gin.WrapH(gwmux))

		gwmuxGroup.Any("/questions", gin.WrapH(gwmux))
		gwmuxGroup.Any("/questions/*path", gin.WrapH(gwmux))

		gwmuxGroup.Any("/users", gin.WrapH(gwmux))
		gwmuxGroup.Any("/users/*path", gin.WrapH(gwmux))

		gwmuxGroup.Any("/history", gin.WrapH(gwmux))
		gwmuxGroup.Any("/history/*path", gin.WrapH(gwmux))
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	})

	return &http.Server{
		Addr:    cfg.HTTP.GetAddr(),
		Handler: router,
	}, nil
}
