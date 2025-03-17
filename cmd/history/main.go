package main

import (
	"fmt"
	"log"
	"os"
	"whoami-server/cmd/history/internal/servers/grpc"
)

func main() {
	config, err := (&grpc.Config{}).Load("configs/default.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	server := grpc.NewServer(config)
	fmt.Printf("Starting History service on port %d\n", config.Server.Port)
	if err := server.RunAndWait(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
		os.Exit(1)
	}
}
