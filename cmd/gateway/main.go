package main

import (
	"fmt"
	"log"
	"os"
	"whoami-server/cmd/gateway/internal/servers/http"
)

func main() {
	config, err := (&http.Config{}).Load("configs/default.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	server := http.NewServer(config)
	fmt.Printf("Starting Whoami service on port %d\n", config.Server.Port)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
		os.Exit(1)
	}
}
