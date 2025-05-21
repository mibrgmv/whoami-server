package auth

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
	"time"
	authpb "whoami-server/protogen/golang/auth"
)

var (
	errAuthHeaderNotProvided = status.Errorf(codes.Unauthenticated, "authorization header is not provided")
	errInvalidAuthFormat     = status.Errorf(codes.Unauthenticated, "invalid authorization header format")
)

const (
	userIDContextKey = "user_id"
)

type Client struct {
	authClient authpb.AuthorizationServiceClient
	conn       *grpc.ClientConn
	timeout    time.Duration
}

func NewClient(authServiceAddr string) (*Client, error) {
	conn, err := grpc.NewClient(authServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := authpb.NewAuthorizationServiceClient(conn)
	return &Client{
		authClient: client,
		conn:       conn,
		timeout:    time.Second * 5,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicEndpoint(r) {
			next.ServeHTTP(w, r)
			return
		}

		token, err := extractToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), c.timeout)
		defer cancel()

		userID, err := c.validateToken(ctx, token)
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				switch st.Code() {
				case codes.Unauthenticated:
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				case codes.Unavailable, codes.DeadlineExceeded:
					http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				}
				return
			}

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		newCtx := context.WithValue(ctx, userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func isPublicEndpoint(r *http.Request) bool {
	publicEndpoints := map[string]string{
		"/api/v1/auth/login":   "*",
		"/api/v1/auth/refresh": "*",
		"/api/v1/quizzes":      "GET",
		"/api/v1/users":        "POST",
	}

	for endpoint, method := range publicEndpoints {
		if r.URL.Path == endpoint && (method == "*" || r.Method == method) {
			return true
		}
	}

	return false
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errAuthHeaderNotProvided
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errInvalidAuthFormat
	}

	return parts[1], nil
}

func (c *Client) validateToken(ctx context.Context, token string) (string, error) {
	resp, err := c.authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		return "", err
	}
	return resp.UserId, nil
}
