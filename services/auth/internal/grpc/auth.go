package grpc

import (
	"context"

	authv1 "github.com/mibrgmv/whoami-server/auth/internal/protogen/auth/v1"
	"github.com/mibrgmv/whoami-server/auth/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authServiceServer struct {
	service service.AuthService
	authv1.UnimplementedAuthServiceServer
}

func NewAuthServiceServer(service service.AuthService) authv1.AuthServiceServer {
	return &authServiceServer{
		service: service,
	}
}

func (h *authServiceServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.TokenResponse, error) {
	accessToken, refreshToken, tokenType, expiresIn, err := h.service.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &authv1.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresIn:    int32(expiresIn),
	}, nil
}

func (h *authServiceServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	userID, username, email, err := h.service.Register(ctx, req.Username, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &authv1.RegisterResponse{
		Id:       userID,
		Username: username,
		Email:    email,
		Message:  "User created successfully",
	}, nil
}

func (h *authServiceServer) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.TokenResponse, error) {
	accessToken, refreshToken, tokenType, expiresIn, err := h.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &authv1.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		ExpiresIn:    int32(expiresIn),
	}, nil
}

func (h *authServiceServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	err := h.service.Logout(ctx, req.RefreshToken)
	if err != nil {
		return nil, h.handleError(err)
	}

	return &authv1.LogoutResponse{
		Message: "Logged out successfully",
	}, nil
}

func (h *authServiceServer) handleError(err error) error {
	switch err {
	case service.ErrInvalidCredentials, service.ErrInvalidToken:
		return status.Error(codes.Unauthenticated, err.Error())
	case service.ErrUsernameExists, service.ErrEmailExists:
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
