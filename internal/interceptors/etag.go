package interceptors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func ETagUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	var ifNoneMatch string
	if ok && len(md["if-none-match"]) > 0 {
		ifNoneMatch = md["if-none-match"][0]
	}

	resp, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}

	if msg, ok := resp.(proto.Message); ok {
		respBytes, err := proto.Marshal(msg)
		if err != nil {
			fmt.Printf("failed to marshal response for ETag: %v\n", err)
			return resp, nil
		}

		hash := sha256.Sum256(respBytes)
		currentETag := fmt.Sprintf("\"%s\"", hex.EncodeToString(hash[:]))

		if ifNoneMatch != "" && (ifNoneMatch == currentETag || ifNoneMatch == "*") {
			err := grpc.SetHeader(ctx, metadata.Pairs("etag", currentETag))
			if err != nil {
				fmt.Printf("failed to set ETag header: %v\n", err)
			}

			err = grpc.SetHeader(ctx, metadata.Pairs("grpc-gateway-http-status", "304"))
			if err != nil {
				fmt.Printf("failed to set grpc-gateway-http-status header: %v\n", err)
			}

			return nil, nil
		}

		if err := grpc.SetHeader(ctx, metadata.Pairs("etag", currentETag)); err != nil {
			fmt.Printf("failed to set ETag header: %v\n", err)
		}
	}

	return resp, nil
}
