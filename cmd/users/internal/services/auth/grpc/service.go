package grpc

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"whoami-server/cmd/users/internal/services/user"
	"whoami-server/internal/tools/jwt"
	pb "whoami-server/protogen/golang/auth"
)

type AuthService struct {
	service *user.Service
	pb.UnimplementedAuthorizationServiceServer
}

func NewAuthService(service *user.Service) *AuthService {
	return &AuthService{
		service: service,
	}
}

func (s *AuthService) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	if request.Username == "" || request.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	userID, err := s.service.Login(ctx, request.Username, request.Password)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		if errors.Is(err, user.ErrIncorrectPassword) {
			return nil, status.Errorf(codes.Unauthenticated, "incorrect password")
		}
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}

	accessToken, refreshToken, err := jwt.GenerateTokenPair(userID.String())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       userID.String(),
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, request *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if request.RefreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	userID, _, err := jwt.ValidateRefreshToken(request.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token: %v", err)
	}

	accessToken, err := jwt.RefreshAccessToken(request.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate new access token: %v", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken: accessToken,
		UserId:      userID,
	}, nil
}
