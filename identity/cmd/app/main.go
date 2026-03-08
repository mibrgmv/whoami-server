package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	appcfg "whoami-server/identity/internal/config"
	"whoami-server/identity/internal/server"
	"whoami-server/libs/config"
)

func main() {
	var cfg appcfg.Config
	var err = config.NewBuilder().
		WithConfigPaths(".").
		WithEnvFiles("../../.env").
		Load(&cfg)

	if err != nil {
		log.Fatalf("failed to read identity service config: %v", err)
	}

	s := server.NewGrpcServer(&cfg)
	lis, err := net.Listen("tcp", cfg.Grpc.GetAddr())
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("gRPC server starting on %s", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC:", err)
		}
	}()

	sig := <-sigCh
	log.Printf("gRPC server shutting down. received signal: %v", sig)
	s.GracefulStop()
}
