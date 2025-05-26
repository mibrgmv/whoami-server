package grpc

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"whoami-server/internal/tools"
)

type Config struct {
	logger *log.Logger
}

func NewConfig(logger *log.Logger) *Config {
	return &Config{
		logger: logger,
	}
}

func authSkip(_ context.Context, c interceptors.CallMeta) bool {
	return c.FullMethod() != "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo" &&
		c.FullMethod() != "/auth.AuthorizationService/Login" &&
		c.FullMethod() != "/auth.AuthorizationService/RefreshToken" &&
		c.FullMethod() != "/auth.AuthorizationService/ValidateToken" &&
		c.FullMethod() != "/user.UserService/CreateUser"
}

func (c *Config) buildLoggingOptions() []logging.Option {
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

func (c *Config) interceptorLogger() logging.Logger {
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
		c.logger.Println(append([]any{"msg", formattedMsg}, fields...)...)
	})
}

func recoveryHandler(p any) (err error) {
	return status.Errorf(codes.Unknown, "panic triggered: %v", p)
}

func (c *Config) buildRecoveryOptions() []recovery.Option {
	return []recovery.Option{
		recovery.WithRecoveryHandler(recoveryHandler),
	}
}

func (c *Config) BuildUnaryInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		logging.UnaryServerInterceptor(c.interceptorLogger(), c.buildLoggingOptions()...),
		selector.UnaryServerInterceptor(tools.MetadataUnaryInterceptor, selector.MatchFunc(authSkip)),
		recovery.UnaryServerInterceptor(c.buildRecoveryOptions()...),
	}
}

func (c *Config) BuildStreamInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		logging.StreamServerInterceptor(c.interceptorLogger(), c.buildLoggingOptions()...),
		selector.StreamServerInterceptor(tools.MetadataStreamInterceptor, selector.MatchFunc(authSkip)),
		recovery.StreamServerInterceptor(c.buildRecoveryOptions()...),
	}
}
