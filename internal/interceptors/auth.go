package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"whoami-server/internal/tools/jwt"
)

var (
	errMissingMetadata       = status.Errorf(codes.InvalidArgument, "missing metadata")
	errAuthHeaderNotProvided = status.Errorf(codes.Unauthenticated, "authorization header is not provided")
	errInvalidAuthFormat     = status.Errorf(codes.Unauthenticated, "invalid authorization format")
	errInvalidToken          = status.Errorf(codes.Unauthenticated, "invalid token")
)

var exemptMethods = map[string]bool{
	"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo": true,

	"/auth.AuthorizationService/Login":        true,
	"/auth.AuthorizationService/RefreshToken": true,

	"/user.UserService/BatchGetUsers":  false,
	"/user.UserService/GetCurrentUser": false,
	"/user.UserService/CreateUser":     true,
	"/user.UserService/UpdateUser":     false,
	"/user.UserService/DeleteUser":     false,

	"/quiz.QuizService/CreateQuiz":      false,
	"/quiz.QuizService/GetQuiz":         false,
	"/quiz.QuizService/BatchGetQuizzes": true,

	"/question.QuestionService/BatchCreateQuestions": false,
	"/question.QuestionService/BatchGetQuestions":    false,
	"/question.QuestionService/EvaluateAnswers":      false,
}

// todo make calls to auth microservice from here -> make a client here
func AuthUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if exemptMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return nil, errAuthHeaderNotProvided
	}

	tokenParts := strings.Split(authHeader[0], " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return nil, errInvalidAuthFormat
	}

	tokenString := tokenParts[1]
	userID, err := jwt.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, errInvalidToken
	}

	newCtx := context.WithValue(ctx, "user_id", userID)

	return handler(newCtx, req)
}

func AuthStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if exemptMethods[info.FullMethod] {
		return handler(srv, ss)
	}

	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return errMissingMetadata
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return errAuthHeaderNotProvided
	}

	tokenParts := strings.Split(authHeader[0], " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return errInvalidAuthFormat
	}

	tokenString := tokenParts[1]
	userID, err := jwt.ValidateAccessToken(tokenString)
	if err != nil {
		return errInvalidToken
	}

	newCtx := context.WithValue(ss.Context(), "user_id", userID)

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
