package grpc

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"whoami-server/cmd/users/internal/services/auth/jwt"
	"whoami-server/cmd/users/internal/services/user"
	pb "whoami-server/protogen/golang/auth"
)

type Service struct {
	service *user.Service
	jwt     *jwt.Service
	pb.UnimplementedAuthorizationServiceServer
}

func NewService(service *user.Service, jwt *jwt.Service) *Service {
	return &Service{
		service: service,
		jwt:     jwt,
	}
}

func (s *Service) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
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

	accessToken, refreshToken, err := s.jwt.GenerateTokenPair(userID.String())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       userID.String(),
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, request *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if request.RefreshToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh token is required")
	}

	userID, _, err := s.jwt.ValidateRefreshToken(request.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token: %v", err)
	}

	accessToken, err := s.jwt.RefreshAccessToken(request.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate new access token: %v", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken: accessToken,
		UserId:      userID,
	}, nil
}

func (s *Service) ValidateToken(ctx context.Context, request *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if request.AccessToken == "" {
		return nil, status.Errorf(codes.InvalidArgument, "access token is required")
	}

	userID, err := s.jwt.ValidateAccessToken(request.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid access token: %v", err)
	}

	return &pb.ValidateTokenResponse{UserId: userID}, nil
}
