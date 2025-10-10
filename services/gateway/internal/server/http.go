package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mibrgmv/whoami-server/gateway/internal/auth"
	"github.com/mibrgmv/whoami-server/gateway/internal/config"
	appcfg "github.com/mibrgmv/whoami-server/gateway/internal/config"
	historypb "github.com/mibrgmv/whoami-server/gateway/internal/protogen/history"
	questionpb "github.com/mibrgmv/whoami-server/gateway/internal/protogen/question"
	quizpb "github.com/mibrgmv/whoami-server/gateway/internal/protogen/quiz"
	userpb "github.com/mibrgmv/whoami-server/gateway/internal/protogen/user"
	"github.com/mibrgmv/whoami-server/shared/keycloak"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func NewServer(ctx context.Context, cfg appcfg.Config) (*gin.Engine, error) {
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

	if err := quizpb.RegisterQuizServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.QuizzesService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register quiz service: %w", err)
	}

	if err := questionpb.RegisterQuestionServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.QuizzesService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register question service: %w", err)
	}

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.UserService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register user service: %w", err)
	}

	if err := historypb.RegisterQuizCompletionHistoryServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.HistoryService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register history service: %w", err)
	}

	keycloakClient := keycloak.NewClient(&cfg.Keycloak)
	authHandler := auth.NewHandler(keycloakClient)
	authMiddleware := auth.NewMiddleware(cfg.Keycloak.BaseURL, cfg.Keycloak.Realm)

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

	router.GET("/api/swagger.json", func(c *gin.Context) {
		c.File("./api/swagger/swagger.json")
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/api/swagger.json")))

	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/register", authHandler.Register)
	}

	gwmuxGroup := router.Group("/api/v1")
	gwmuxGroup.Use(authMiddleware.Authorization())
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

	return router, nil
}

func StartHTTP(ctx context.Context, cfg config.Config) error {
	router, err := NewServer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create HTTP server: %w", err)
	}

	server := &http.Server{
		Addr:    cfg.HTTP.GetAddr(),
		Handler: router,
	}

	log.Println("Starting HTTP server on", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
