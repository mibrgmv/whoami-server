package grpc

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"whoami-server/cmd/users/internal/models"
	"whoami-server/cmd/users/internal/services/user"
	"whoami-server/internal/jwt"
	pb "whoami-server/protogen/golang/user"
)

type UserService struct {
	service *user.Service
	pb.UnimplementedUserServiceServer
}

func NewUserService(service *user.Service) *UserService {
	return &UserService{
		service: service,
	}
}

func (s *UserService) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.User, error) {
	if request.Username == "" || request.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	usr, err := s.service.Register(ctx, request.Username, request.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &pb.User{
		UserId:    usr.ID.String(),
		Username:  usr.Name,
		Password:  usr.Password,
		CreatedAt: timestamppb.New(usr.CreatedAt),
		LastLogin: timestamppb.New(usr.LastLogin),
	}, nil
}

func (s *UserService) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
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

func (s *UserService) Update(ctx context.Context, request *pb.UpdateRequest) (*pb.User, error) {
	userID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	usrPtr, err := s.service.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}
	usr := *usrPtr

	if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(request.CurrentPassword)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect password")
	}

	if request.NewPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password")
		}
		usr.Password = string(hashedPassword)
	}

	// todo username change does not work
	updatedUsers, err := s.service.Update(ctx, []*models.User{&usr})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	updatedUser := updatedUsers[0]

	return &pb.User{
		UserId:    updatedUser.ID.String(),
		Username:  updatedUser.Name,
		Password:  updatedUser.Password,
		CreatedAt: timestamppb.New(updatedUser.CreatedAt),
		LastLogin: timestamppb.New(updatedUser.LastLogin),
	}, nil
}

func (s *UserService) Delete(ctx context.Context, request *pb.DeleteRequest) (*emptypb.Empty, error) {
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	requesterID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid user ID format: %v", err)
	}

	targetUserID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target user ID format: %v", err)
	}

	if requesterID != targetUserID {
		return nil, status.Errorf(codes.PermissionDenied, "you are not authorized to delete this user")
	}

	err = s.service.Delete(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *UserService) RefreshToken(ctx context.Context, request *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
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

func (s *UserService) GetCurrent(ctx context.Context, empty *emptypb.Empty) (*pb.User, error) {
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	usr, err := s.service.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return usr.ToProto(), nil
}

func (s *UserService) GetBatch(ctx context.Context, request *pb.GetBatchRequest) (*pb.GetBatchResponse, error) {
	userIDStr, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	_, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	users, nextPageToken, err := s.service.Get(ctx, request.PageSize, request.PageToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get users: %v", err)
	}

	var pbUsers []*pb.User
	for _, u := range users {
		pbUsers = append(pbUsers, u.ToProto())
	}

	return &pb.GetBatchResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
	}, nil
}
