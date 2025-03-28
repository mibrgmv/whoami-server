package grpc

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
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
		UserId:    usr.ID,
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
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}

	tokenString, err := jwt.GenerateToken(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &pb.LoginResponse{
		Token:  tokenString,
		UserId: userID,
	}, nil
}

func (s *UserService) GetCurrent(ctx context.Context, empty *emptypb.Empty) (*pb.User, error) {
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	usr, err := s.service.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return models.ToProto(usr), nil
}

func (s *UserService) GetAll(empty *emptypb.Empty, stream pb.UserService_GetAllServer) error {
	_, ok := stream.Context().Value("user_id").(int64)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	users, err := s.service.GetAll(stream.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "failed to get users: %v", err)
	}

	done := make(chan bool)

	go func() {
		defer close(done)
		for _, u := range users {
			if err := stream.Send(models.ToProto(&u)); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
	}()

	<-done
	return nil
}
