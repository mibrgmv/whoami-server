.PHONY: lint gen swag build push

MODULE=github.com/mibrgmv/whoami-server/gateway
PROTO_DIR=api/v1
PROTOGEN_DIR=internal/protogen
THIRD_PARTY_DIR=../../shared/third_party

lint:
	golangci-lint run

gen:
	protoc -I=$(PROTO_DIR) -I=$(THIRD_PARTY_DIR) \
		--go_out=. \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(MODULE) \
		--grpc-gateway_out=. \
		--grpc-gateway_opt=module=$(MODULE) \
		--openapiv2_out=$(PROTO_DIR) \
		--openapiv2_opt=allow_merge=true,merge_file_name=gateway \
		$(PROTO_DIR)/*.proto
