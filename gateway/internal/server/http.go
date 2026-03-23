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
	goredis "github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"gordle/gateway/internal/auth"
	appcfg "gordle/gateway/internal/config"
	"gordle/gateway/internal/handler"
	"gordle/gateway/internal/metrics"
	"gordle/gateway/internal/middleware"
	redisrepo "gordle/gateway/internal/repository/redis"
	"gordle/gateway/internal/websocket"
	gamev1 "gordle/gateway/pkg/protogen/game/v1"
	roomv1 "gordle/gateway/pkg/protogen/room/v1"
	statisticsv1 "gordle/gateway/pkg/protogen/statistics/v1"
	"gordle/libs/keycloak"
)

func NewHttpServer(ctx context.Context, cfg appcfg.Config, collector *metrics.Collector, logger *slog.Logger, redisClient *goredis.Client, keycloakClient *keycloak.Client) (*http.Server, error) {
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
		{"game", gamev1.RegisterGameServiceHandlerFromEndpoint, cfg.GameService.GetAddr()},
		{"room", roomv1.RegisterRoomServiceHandlerFromEndpoint, cfg.GameService.GetAddr()},
		{"statistics", statisticsv1.RegisterStatisticsServiceHandlerFromEndpoint, cfg.StatisticsService.GetAddr()},
	}

	for _, svc := range services {
		if err := svc.register(ctx, gwmux, svc.addr, dialOpts); err != nil {
			return nil, fmt.Errorf("failed to register %s server service: %w", svc.name, err)
		}
	}

	sessionRepo := redisrepo.NewSessionRepository(redisClient)

	keycloakValidator := auth.NewKeycloakValidator(auth.KeycloakValidatorConfig{
		BaseURL:       cfg.Keycloak.BaseURL,
		IssuerURL:     cfg.Keycloak.IssuerURL,
		Realm:         cfg.Keycloak.Realm,
		KeyRefreshTTL: 1 * time.Hour,
		HTTPTimeout:   10 * time.Second,
	})

	var guestValidator *auth.GuestTokenValidator
	if cfg.GuestSecret != "" {
		guestValidator = auth.NewGuestTokenValidator(cfg.GuestSecret)
	}

	authConfig := middleware.AuthConfig{
		KeycloakValidator: keycloakValidator,
		GuestValidator:    guestValidator,
		SessionRepo:       sessionRepo,
		KeycloakClient:    keycloakClient,
		CookieName:        cfg.Session.CookieName,
		Metrics:           collector,
	}
	authMiddleware := middleware.Auth(authConfig)
	authOptionalMiddleware := middleware.AuthOptional(authConfig)

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

	authHandler := handler.NewAuthHandler(sessionRepo, keycloakClient, cfg.Session)
	router.GET("/api/v1/auth/login", authHandler.Login)
	router.GET("/api/v1/auth/register", authHandler.Register)
	router.GET("/api/v1/auth/callback", authHandler.Callback)
	router.GET("/api/v1/auth/logout", authHandler.Logout)
	router.GET("/api/v1/auth/me", authOptionalMiddleware, authHandler.Me)
	router.GET("/api/v1/auth/action", authHandler.Action)

	router.POST("/api/v1/auth/guest", handler.GuestAuth(cfg.GuestSecret))

	gamesGroup := router.Group("/api/v1")
	gamesGroup.Use(authOptionalMiddleware)
	{
		gamesGroup.Any("/games", gin.WrapH(gwmux))
		gamesGroup.Any("/games/*path", gin.WrapH(gwmux))
	}

	roomsGroup := router.Group("/api/v1/rooms")
	{
		roomsGroup.POST("", authMiddleware, gin.WrapH(gwmux))

		roomsGroup.Use(authOptionalMiddleware)
		roomsGroup.GET("/:code", gin.WrapH(gwmux))
		roomsGroup.POST("/:code/join", gin.WrapH(gwmux))

		if cfg.GameWebSocket.Port > 0 {
			wsProxy, err := websocket.NewProxy(cfg.GameWebSocket.GetAddr(), logger)
			if err != nil {
				logger.Error("failed to create WebSocket proxy", slog.String("error", err.Error()))
			} else {
				roomsGroup.GET("/:code/ws", gin.WrapH(wsProxy))
			}
		}
	}

	protectedGroup := router.Group("/api/v1")
	protectedGroup.Use(authMiddleware)
	{
		protectedGroup.Any("/statistics", gin.WrapH(gwmux))
		protectedGroup.Any("/statistics/*path", gin.WrapH(gwmux))
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	})

	return &http.Server{
		Addr:    cfg.HTTP.GetAddr(),
		Handler: router,
	}, nil
}
