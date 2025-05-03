package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var cacheControlMethods = map[string]string{
	"/user.UserService/GetBatch": "public, max-age=300",
}

func CacheControlUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if cacheControlHeader, ok := cacheControlMethods[info.FullMethod]; ok {
		newCtx := metadata.AppendToOutgoingContext(ctx, "cache-control", cacheControlHeader)
		return handler(newCtx, req)
	}
	return handler(ctx, req)
}

func CacheControlStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if cacheControlHeader, ok := cacheControlMethods[info.FullMethod]; ok {
		wrappedStream := &ctxServerStream{
			ServerStream: ss,
			ctx:          metadata.AppendToOutgoingContext(ss.Context(), "cache-control", cacheControlHeader),
		}

		if err := ss.SendHeader(metadata.Pairs("cache-control", cacheControlHeader)); err != nil {
			return err
		}

		return handler(srv, wrappedStream)
	}
	return handler(srv, ss)
}
