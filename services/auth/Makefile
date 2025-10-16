.PHONY: lint gen

lint:
	golangci-lint run

gen:
	protoc -I=api/v1 -I=../../shared/third_party \
		--go_out=. \
		--go_opt=module=github.com/mibrgmv/whoami-server/auth \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/mibrgmv/whoami-server/auth \
		api/v1/*.proto