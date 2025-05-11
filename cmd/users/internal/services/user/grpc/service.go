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

func (s *UserService) BatchGetUsers(ctx context.Context, request *pb.BatchGetUsersRequest) (*pb.BatchGetUsersResponse, error) {
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

	return &pb.BatchGetUsersResponse{
		Users:         pbUsers,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *UserService) GetCurrentUser(ctx context.Context, empty *emptypb.Empty) (*pb.User, error) {
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

func (s *UserService) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.User, error) {
	if request.User.Username == "" || request.User.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	usr, err := s.service.Register(ctx, request.User.Username, request.User.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &pb.User{
		UserId:    usr.ID.String(),
		Username:  usr.Name,
		Password:  "",
		CreatedAt: timestamppb.New(usr.CreatedAt),
		LastLogin: timestamppb.New(usr.LastLogin),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, request *pb.UpdateUserRequest) (*pb.User, error) {
	userID, err := uuid.Parse(request.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID format: %v", err)
	}

	existingUser, err := s.service.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(request.CurrentPassword)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "incorrect password")
	}
	// todo test field mask
	// todo move some logic to BL
	updatedUser := *existingUser

	updateAll := len(request.UpdateMask.GetPaths()) == 0

	for _, path := range request.UpdateMask.GetPaths() {
		switch path {
		case "username":
			if request.User.Username != "" {
				updatedUser.Name = request.User.Username
			}
		case "password":
			if request.User.Password != "" {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.User.Password), bcrypt.DefaultCost)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "failed to hash password")
				}
				updatedUser.Password = string(hashedPassword)
			}
		default:
			return nil, status.Errorf(codes.InvalidArgument, "update mask contains invalid field: %s", path)
		}
	}

	if updateAll {
		if request.User.Username != "" {
			updatedUser.Name = request.User.Username
		}
		if request.User.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.User.Password), bcrypt.DefaultCost)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to hash password")
			}
			updatedUser.Password = string(hashedPassword)
		}
	}

	updatedUsers, err := s.service.Update(ctx, []*models.User{&updatedUser})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	updatedUser = *updatedUsers[0]

	return &pb.User{
		UserId:    updatedUser.ID.String(),
		Username:  updatedUser.Name,
		Password:  "",
		CreatedAt: timestamppb.New(updatedUser.CreatedAt),
		LastLogin: timestamppb.New(updatedUser.LastLogin),
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, request *pb.DeleteUserRequest) (*emptypb.Empty, error) {
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

	err = s.service.Delete(ctx, &targetUserID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &emptypb.Empty{}, nil
}
