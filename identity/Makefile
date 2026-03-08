.PHONY: lint gen build

build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/app ./cmd/*/

lint:
	golangci-lint run

gen:
	protoc -I=api/v1 -I=../libs/third_party \
		--go_out=. \
		--go_opt=module=whoami-server/identity \
		--go-grpc_out=. \
		--go-grpc_opt=module=whoami-server/identity \
		api/v1/*.proto
