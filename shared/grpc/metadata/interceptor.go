package metadata

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strconv"
)

const (
	UserIDKey        string = "user_id"
	UsernameKey      string = "username"
	EmailKey         string = "email"
	EmailVerifiedKey string = "email_verified"
)

func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := enrichContextFromMetadata(ctx)

		return handler(newCtx, req)
	}
}

func StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := enrichContextFromMetadata(ss.Context())

		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

func enrichContextFromMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	newCtx := ctx

	if userIDs := md.Get("user_id"); len(userIDs) > 0 {
		newCtx = context.WithValue(newCtx, UserIDKey, userIDs[0])
	}

	if usernames := md.Get("username"); len(usernames) > 0 {
		newCtx = context.WithValue(newCtx, UsernameKey, usernames[0])
	}

	if emails := md.Get("email"); len(emails) > 0 {
		newCtx = context.WithValue(newCtx, EmailKey, emails[0])
	}

	if emailVerified := md.Get("email_verified"); len(emailVerified) > 0 {
		verified, _ := strconv.ParseBool(emailVerified[0])
		newCtx = context.WithValue(newCtx, EmailVerifiedKey, verified)
	}

	return newCtx
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDStr, ok := ctx.Value(UserIDKey).(string)
	if !ok || userIDStr == "" {
		return uuid.UUID{}, fmt.Errorf("user_id not found in context")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid user_id format: %v", err)
	}

	return userID, nil
}

func GetUsernameFromContext(ctx context.Context) (string, error) {
	username, ok := ctx.Value(UsernameKey).(string)
	if !ok || username == "" {
		return "", fmt.Errorf("username not found in context")
	}
	return username, nil
}

func GetEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(EmailKey).(string)
	if !ok || email == "" {
		return "", fmt.Errorf("email not found in context")
	}
	return email, nil
}
