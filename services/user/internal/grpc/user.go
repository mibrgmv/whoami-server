package grpc

import (
	"context"

	userpb "github.com/mibrgmv/whoami-server/user/internal/protogen/user"
	"github.com/mibrgmv/whoami-server/user/internal/service"
	"github.com/mibrgmv/whoami-server/user/internal/service/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userServiceServer struct {
	service service.UserService
	userpb.UnimplementedUserServiceServer
}

func NewUserServiceServer(service service.UserService) userpb.UserServiceServer {
	return &userServiceServer{
		service: service,
	}
}

func (s *userServiceServer) GetCurrentUser(ctx context.Context, empty *emptypb.Empty) (*userpb.User, error) {
	userID, _ := ctx.Value("user_id").(string)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	user, err := s.service.GetUser(ctx, userID)
	if err != nil {
		return nil, s.handleError(err)
	}

	return user.ToProto(), nil
}

func (s *userServiceServer) BatchGetUsers(ctx context.Context, req *userpb.BatchGetUsersRequest) (*userpb.BatchGetUsersResponse, error) {
	if _, ok := ctx.Value("user_id").(string); !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	users, nextOffset, err := s.service.BatchGetUsers(ctx, req.PageSize, req.Offset)
	if err != nil {
		return nil, s.handleError(err)
	}

	pbUsers := make([]*userpb.User, len(users))
	for i, user := range users {
		pbUsers[i] = user.ToProto()
	}

	return &userpb.BatchGetUsersResponse{
		Users:      pbUsers,
		NextOffset: nextOffset,
	}, nil
}

func (s *userServiceServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.User, error) {
	authUserID, ok := ctx.Value("user_id").(string)
	if !ok || authUserID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if authUserID != req.Id {
		return nil, status.Error(codes.PermissionDenied, "not authorized to update this user")
	}

	updateData := models.UpdateUserData{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	user, err := s.service.UpdateUser(ctx, req.Id, updateData)
	if err != nil {
		return nil, s.handleError(err)
	}

	return user.ToProto(), nil
}

func (s *userServiceServer) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*userpb.ChangePasswordResponse, error) {
	authUserID, ok := ctx.Value("user_id").(string)
	if !ok || authUserID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if authUserID != req.Id {
		return nil, status.Error(codes.PermissionDenied, "not authorized to change this user's password")
	}

	err := s.service.ChangePassword(ctx, req.Id, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return nil, s.handleError(err)
	}

	return &userpb.ChangePasswordResponse{
		Message: "Password updated successfully",
	}, nil
}

func (s *userServiceServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	authUserID, ok := ctx.Value("user_id").(string)
	if !ok || authUserID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if authUserID != req.Id {
		return nil, status.Error(codes.PermissionDenied, "not authorized to delete this user")
	}

	err := s.service.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, s.handleError(err)
	}

	return &userpb.DeleteUserResponse{
		Id:      req.Id,
		Message: "User deleted successfully",
	}, nil
}

func (s *userServiceServer) handleError(err error) error {
	switch err {
	case service.ErrUserNotFound:
		return status.Error(codes.NotFound, err.Error())
	case service.ErrUsernameExists, service.ErrEmailExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case service.ErrInvalidPassword, service.ErrPasswordPolicy:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
