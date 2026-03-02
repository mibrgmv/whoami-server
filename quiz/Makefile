.PHONY: lint gen


lint:
	golangci-lint run

gen:
	protoc -I=api/v1 -I=../libs/third_party \
		--go_out=. \
		--go_opt=module=whoami-server/quiz \
		--go-grpc_out=. \
		--go-grpc_opt=module=whoami-server/quiz \
		api/v1/*.proto