package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"strings"
	"whoami-server-gateway/internal/auth"
	"whoami-server-gateway/internal/auth/keycloak"
	"whoami-server-gateway/internal/config"
	appcfg "whoami-server-gateway/internal/config"
	auth2 "whoami-server-gateway/internal/handlers/auth"
	"whoami-server-gateway/internal/servers/http/cors"
	historypb "whoami-server-gateway/protogen/golang/history"
	questionpb "whoami-server-gateway/protogen/golang/question"
	quizpb "whoami-server-gateway/protogen/golang/quiz"
	userpb "whoami-server-gateway/protogen/golang/user"
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

			if username := auth.GetUsername(ctx); username != "" {
				md.Set("username", username)
			}

			if email := auth.GetEmail(ctx); email != "" {
				md.Set("email", email)
			}

			return md
		}),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		cfg.UsersService.GetAddr(),
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register user service: %w", err)
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

	keycloakClient := keycloak.NewClient(cfg.Keycloak)
	authHandler := auth2.NewHandler(keycloakClient, nil)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(
		auth.GinJWTMiddleware(cfg.Keycloak.BaseURL, cfg.Keycloak.Realm),
		cors.Middleware(cfg.HTTP.CORS),
		gin.Logger(),
		gin.Recovery(),
	)

	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/register", authHandler.Register)
	}

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/") &&
			!strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth/") {
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

	//handler := ApplyMiddleware(router,
	//	auth.NewJWTMiddleware(cfg.Keycloak.BaseURL, cfg.Keycloak.Realm),
	//	cors.Middleware(cfg.HTTP.CORS),
	//	logging.Middleware,
	//	recovery.Middleware,
	//)

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

func ApplyMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
