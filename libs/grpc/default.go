package grpc

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func DefaultUnaryInterceptors(logger *slog.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		UnaryMetadataInterceptor(),
		logging.UnaryServerInterceptor(slogAdapter(logger), loggingOptions()...),
		recovery.UnaryServerInterceptor(recoveryOptions()...),
	}
}

func DefaultStreamInterceptors(logger *slog.Logger) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		StreamMetadataInterceptor(),
		logging.StreamServerInterceptor(slogAdapter(logger), loggingOptions()...),
		recovery.StreamServerInterceptor(recoveryOptions()...),
	}
}

func slogAdapter(logger *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l := logger

		if requestID := GetRequestIDFromContext(ctx); requestID != "" {
			l = l.With(slog.String("request_id", requestID))
		}
		if userID, _ := ctx.Value(UserIDKey).(string); userID != "" {
			l = l.With(slog.String("user_id", userID))
		}

		l = l.With(fields...)

		switch lvl {
		case logging.LevelDebug:
			l.DebugContext(ctx, msg)
		case logging.LevelInfo:
			l.InfoContext(ctx, msg)
		case logging.LevelWarn:
			l.WarnContext(ctx, msg)
		case logging.LevelError:
			l.ErrorContext(ctx, msg)
		}
	})
}

func loggingOptions() []logging.Option {
	return []logging.Option{
		logging.WithLogOnEvents(
			logging.StartCall,
			logging.FinishCall,
		),
		logging.WithLevels(func(code codes.Code) logging.Level {
			switch code {
			case codes.OK:
				return logging.LevelInfo
			case codes.Canceled, codes.InvalidArgument, codes.NotFound,
				codes.AlreadyExists, codes.PermissionDenied, codes.Unauthenticated:
				return logging.LevelInfo
			case codes.DeadlineExceeded, codes.ResourceExhausted, codes.FailedPrecondition,
				codes.Aborted, codes.OutOfRange, codes.Unavailable:
				return logging.LevelWarn
			case codes.Unknown, codes.Internal, codes.DataLoss:
				return logging.LevelError
			default:
				return logging.LevelError
			}
		}),
	}
}

func recoveryOptions() []recovery.Option {
	return []recovery.Option{
		recovery.WithRecoveryHandler(recoveryHandler),
	}
}

func recoveryHandler(p any) error {
	return status.Errorf(codes.Internal, "internal server error: %v", p)
}
