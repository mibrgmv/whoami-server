package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	pb "whoami-server/protogen/golang/user"
)

func NewServer(ctx context.Context, grpcAddr string) (*gin.Engine, error) {
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

	err := pb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		grpcAddr,
		dialOpts,
	)
	if err != nil {
		return nil, err
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	r.Any("/api/v1/*any", func(c *gin.Context) {
		gwmux.ServeHTTP(c.Writer, c.Request)
	})

	return r, nil
}

func Start(ctx context.Context, grpcAddr, httpAddr string) error {
	r, err := NewServer(ctx, grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to create HTTP server: %w", err)
	}

	log.Println("Serving gRPC-Gateway on", httpAddr)
	if err := r.Run(httpAddr); err != nil {
		return fmt.Errorf("failed to run HTTP server: %w", err)
	}

	return nil
}
