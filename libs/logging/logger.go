package logging

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

func NewLogger(serviceName string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return slog.New(handler).With(slog.String("service", serviceName))
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func FromContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(slog.String("request_id", requestID))
	}

	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		logger = logger.With(slog.String("user_id", userID))
	}

	return logger
}
