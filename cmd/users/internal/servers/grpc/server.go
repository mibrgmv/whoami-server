package grpc

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"whoami-server/cmd/users/internal/services/user"
	usergrpc "whoami-server/cmd/users/internal/services/user/grpc"
	pg "whoami-server/cmd/users/internal/services/user/postgresql"
	"whoami-server/internal/jwt"
	userpb "whoami-server/protogen/golang/user"
)

func NewServer(pool *pgxpool.Pool) *grpc.Server {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(jwt.AuthUnaryInterceptor),
		grpc.StreamInterceptor(jwt.AuthStreamInterceptor),
	)
	repo := pg.NewRepository(pool)
	service := user.NewService(repo)
	server := usergrpc.NewUserService(service)
	userpb.RegisterUserServiceServer(s, server)
	reflection.Register(s)
	return s
}

func Start(pool *pgxpool.Pool, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	s := NewServer(pool)
	log.Println("Serving gRPC on", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}
