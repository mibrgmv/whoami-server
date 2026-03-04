.PHONY: lint gen swag build push

GO_MODULE=whoami-server/gateway
PROTO_DIR=api/v1

build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/app ./cmd/*/

lint:
	golangci-lint run

gen:
	protoc -I=$(PROTO_DIR) -I=../libs/third_party \
		--go_out=. \
		--go_opt=module=$(GO_MODULE) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(GO_MODULE) \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=module=$(GO_MODULE) \
		--openapiv2_out=$(PROTO_DIR) \
		--openapiv2_opt=allow_merge=true,merge_file_name=gateway \
		$(PROTO_DIR)/*.proto
