package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"net/http"
	"whoami-server/cmd/gateway/internal/servers/http/auth"
	"whoami-server/cmd/gateway/internal/servers/http/cors"
	"whoami-server/cmd/gateway/internal/servers/http/logging"
	"whoami-server/cmd/gateway/internal/servers/http/recovery"
	httpcfg "whoami-server/internal/config/api/http"
	authpb "whoami-server/protogen/golang/auth"
	historypb "whoami-server/protogen/golang/history"
	questionpb "whoami-server/protogen/golang/question"
	quizpb "whoami-server/protogen/golang/quiz"
	userpb "whoami-server/protogen/golang/user"
)

func NewServer(ctx context.Context, grpcAddresses map[string]string) (*http.ServeMux, error) {
	gwmux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			md := metadata.New(map[string]string{
				"authorization": req.Header.Get("Authorization"),
			})

			if userId, ok := ctx.Value("user_id").(string); ok && userId != "" {
				md.Set("user_id", userId)
			}

			return md
		}),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := authpb.RegisterAuthorizationServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		grpcAddresses["users"],
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register user service: %w", err)
	}

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		grpcAddresses["users"],
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register user service: %w", err)
	}

	if err := quizpb.RegisterQuizServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		grpcAddresses["quizzes"],
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register quiz service: %w", err)
	}

	if err := questionpb.RegisterQuestionServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		grpcAddresses["quizzes"],
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register question service: %w", err)
	}

	if err := historypb.RegisterQuizCompletionHistoryServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		grpcAddresses["history"],
		dialOpts,
	); err != nil {
		return nil, fmt.Errorf("failed to register history service: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/v1/", gwmux)

	return mux, nil
}

func Start(ctx context.Context, grpcAddresses map[string]string, httpConfig httpcfg.Config) error {
	mux, err := NewServer(ctx, grpcAddresses)
	if err != nil {
		return fmt.Errorf("failed to create HTTP server: %w", err)
	}

	authClient, err := auth.NewClient(grpcAddresses["users"])
	if err != nil {
		return fmt.Errorf("failed to create auth middleware: %w", err)
	}
	defer authClient.Close()

	handler := ApplyMiddleware(mux,
		authClient.Middleware,
		cors.Middleware(httpConfig.CORS),
		logging.Middleware,
		recovery.Middleware,
	)

	server := &http.Server{
		Addr:    httpConfig.GetAddr(),
		Handler: handler,
	}

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
