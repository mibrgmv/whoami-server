package interceptor

import (
	"context"
	"fmt"
	"log"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func DefaultUnaryInterceptors(logger *log.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		logging.UnaryServerInterceptor(loggingAdapter(logger), loggingOptions()...),
		recovery.UnaryServerInterceptor(recoveryOptions()...),
	}
}

func DefaultStreamInterceptors(logger *log.Logger) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		logging.StreamServerInterceptor(loggingAdapter(logger), loggingOptions()...),
		recovery.StreamServerInterceptor(recoveryOptions()...),
	}
}

func loggingAdapter(logger *log.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		var prefix string
		switch lvl {
		case logging.LevelDebug:
			prefix = "DEBUG"
		case logging.LevelInfo:
			prefix = "INFO"
		case logging.LevelWarn:
			prefix = "WARN"
		case logging.LevelError:
			prefix = "ERROR"
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}

		formattedMsg := fmt.Sprintf("%s: %v", prefix, msg)
		logger.Println(append([]any{"msg", formattedMsg}, fields...)...)
	})
}

func loggingOptions() []logging.Option {
	return []logging.Option{
		logging.WithLogOnEvents(
			logging.StartCall,
			logging.FinishCall,
			logging.PayloadReceived,
			logging.PayloadSent,
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
