package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"whoami-server/cmd/users/internal/services/user"
	usergrpc "whoami-server/cmd/users/internal/services/user/grpc"
	pg "whoami-server/cmd/users/internal/services/user/postgresql"
	"whoami-server/internal/jwt"
	userpb "whoami-server/protogen/golang/user"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	connString := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=prefer"
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse connStr: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		log.Fatalf("Unable to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	s := grpc.NewServer(
		grpc.UnaryInterceptor(jwt.AuthUnaryInterceptor),
		grpc.StreamInterceptor(jwt.AuthStreamInterceptor),
	)
	repo := pg.NewRepository(pool)
	service := user.NewService(repo)
	server := usergrpc.NewUserService(service)
	userpb.RegisterUserServiceServer(s, server)
	reflection.Register(s)
	log.Println("Serving gRPC on 0.0.0.0:8080")
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	gwmux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			return metadata.New(map[string]string{
				"authorization": req.Header.Get("Authorization"),
			})
		}),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err = userpb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		"localhost:8080",
		dialOpts,
	)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		gwmux.ServeHTTP(w, r)
	})

	httpServer := &http.Server{
		Addr:    ":8090",
		Handler: handler,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	log.Fatalln(httpServer.ListenAndServe())
}
