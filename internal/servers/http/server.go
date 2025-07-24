package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"strings"
	"whoami-server-gateway/api/swagger"
	"whoami-server-gateway/internal/auth"
	"whoami-server-gateway/internal/auth/keycloak"
	"whoami-server-gateway/internal/config"
	appcfg "whoami-server-gateway/internal/config"
	"whoami-server-gateway/internal/users"
	historypb "whoami-server-gateway/protogen/golang/history"
	questionpb "whoami-server-gateway/protogen/golang/question"
	quizpb "whoami-server-gateway/protogen/golang/quiz"
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

	if err := historypb.RegisterQuizCompletionHistoryServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.HistoryService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register history service: %w", err)
	}

	keycloakClient := keycloak.NewClient(&cfg.Keycloak)
	authHandler := auth.NewHandler(keycloakClient, nil)
	authMiddleware := auth.NewMiddleware(cfg.Keycloak.BaseURL, cfg.Keycloak.Realm)
	usersHandler := users.NewHandler(keycloakClient, nil)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	swagger.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/register", authHandler.Register)
	}

	usersGroup := router.Group("/api/v1/users")
	usersGroup.Use(authMiddleware.Authorization())
	{
		usersGroup.GET("/current", usersHandler.GetCurrentUser)
		usersGroup.GET("", usersHandler.BatchGetUsers)
		usersGroup.PUT("/:id", usersHandler.UpdateUser)
		usersGroup.PUT("/:id/password", usersHandler.ChangePassword)
		usersGroup.DELETE("/:id", usersHandler.DeleteUser)
	}

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/") &&
			!strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth/") {
			authMiddleware.Authorization()
			gin.WrapH(gwmux)(c)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		}
	})

	return router, nil
}

func Start(ctx context.Context, cfg config.Config) error {
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
