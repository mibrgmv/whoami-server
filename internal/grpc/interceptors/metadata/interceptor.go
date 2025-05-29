package metadata

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errMissingUserID   = status.Errorf(codes.InvalidArgument, "missing userID from metadata")
)

const (
	userIDContextKey string = "user_id"
)

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}

	userIDValues := md.Get("user_id")
	if len(userIDValues) == 0 || userIDValues[0] == "" {
		return nil, errMissingUserID
	}

	userID := userIDValues[0]
	newCtx := context.WithValue(ctx, userIDContextKey, userID)
	return handler(newCtx, req)
}

func StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return errMissingMetadata
	}

	userIDValues := md.Get("user_id")
	if len(userIDValues) == 0 || userIDValues[0] == "" {
		return errMissingUserID
	}

	userID := userIDValues[0]
	newCtx := context.WithValue(ss.Context(), userIDContextKey, userID)

	wrappedStream := wrapServerStream(ss)
	wrappedStream.SetContext(newCtx)
	return handler(srv, wrappedStream)
}

func wrapServerStream(ss grpc.ServerStream) *wrappedServerStream {
	return &wrappedServerStream{ss}
}

type wrappedServerStream struct {
	grpc.ServerStream
}

func (w *wrappedServerStream) SetContext(ctx context.Context) {
	w.ServerStream = &ctxServerStream{ctx, w.ServerStream}
}

type ctxServerStream struct {
	ctx context.Context
	grpc.ServerStream
}

func (c *ctxServerStream) Context() context.Context {
	return c.ctx
}
