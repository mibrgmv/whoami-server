package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	appcfg "whoami-server/gateway/internal/config"
	"whoami-server/gateway/internal/metrics"
	"whoami-server/gateway/internal/middleware"
	authv1 "whoami-server/gateway/pkg/protogen/auth/v1"
	historyv1 "whoami-server/gateway/pkg/protogen/history/v1"
	questionv1 "whoami-server/gateway/pkg/protogen/question/v1"
	quizv1 "whoami-server/gateway/pkg/protogen/quiz/v1"
	userv1 "whoami-server/gateway/pkg/protogen/user/v1"
)

func NewHttpServer(ctx context.Context, cfg appcfg.Config, collector *metrics.Collector, logger *slog.Logger) (*http.Server, error) {
	gwmux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			md := metadata.New(map[string]string{
				"authorization": req.Header.Get("Authorization"),
			})

			reqCtx := req.Context()
			if userId, ok := reqCtx.Value(middleware.UserIDKey).(string); ok && userId != "" {
				md.Set("user_id", userId)
			}

			if username, ok := reqCtx.Value(middleware.UsernameKey).(string); ok && username != "" {
				md.Set("username", username)
			}

			if email, ok := reqCtx.Value(middleware.EmailKey).(string); ok && email != "" {
				md.Set("email", email)
			}

			if roles, ok := reqCtx.Value(middleware.RolesKey).(string); ok && roles != "" {
				md.Set("roles", roles)
			}

			if requestID, ok := reqCtx.Value(middleware.RequestIDKey).(string); ok && requestID != "" {
				md.Set("request_id", requestID)
			}

			return md
		}),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	services := []struct {
		name     string
		register func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
		addr     string
	}{
		{"auth", authv1.RegisterAuthServiceHandlerFromEndpoint, cfg.AuthService.GetAddr()},
		{"quiz", quizv1.RegisterQuizServiceHandlerFromEndpoint, cfg.QuizService.GetAddr()},
		{"question", questionv1.RegisterQuestionServiceHandlerFromEndpoint, cfg.QuizService.GetAddr()},
		{"user", userv1.RegisterUserServiceHandlerFromEndpoint, cfg.UserService.GetAddr()},
		{"history", historyv1.RegisterHistoryServiceHandlerFromEndpoint, cfg.HistoryService.GetAddr()},
	}

	for _, svc := range services {
		if err := svc.register(ctx, gwmux, svc.addr, dialOpts); err != nil {
			return nil, fmt.Errorf("failed to register %s server service: %w", svc.name, err)
		}
	}

	jwtMiddleware := middleware.JWT(middleware.JWTConfig{
		KeycloakBaseURL: cfg.Keycloak.BaseURL,
		Realm:           cfg.Keycloak.Realm,
		KeyRefreshTTL:   1 * time.Hour,
		HTTPTimeout:     10 * time.Second,
		Metrics:         collector,
	})

	switch cfg.HTTP.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
		logger.Info("running in debug mode")
		if configJSON, err := json.MarshalIndent(cfg, "", "  "); err == nil {
			logger.Debug("configuration", slog.String("config", string(configJSON)))
		}
	case "release":
		gin.SetMode(gin.ReleaseMode)
		logger.Info("running in release mode")
	default:
		gin.SetMode(gin.ReleaseMode)
		logger.Warn("unknown mode, defaulting to release", slog.String("mode", cfg.HTTP.Mode))
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.HTTP.CORS.AllowedOrigins,
		AllowMethods:     cfg.HTTP.CORS.AllowedMethods,
		AllowHeaders:     cfg.HTTP.CORS.AllowedHeaders,
		ExposeHeaders:    cfg.HTTP.CORS.ExposeHeaders,
		AllowCredentials: cfg.HTTP.CORS.AllowCredentials,
		MaxAge:           cfg.HTTP.CORS.MaxAge,
	}))

	router.Use(metrics.GinMiddleware(collector))
	router.Use(middleware.RequestID())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

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
