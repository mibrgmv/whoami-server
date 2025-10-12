.PHONY: lint gen

lint:
	golangci-lint run

gen:
	@mkdir -p internal/protogen/history
	protoc -I=api/v1 -I=../../shared/third_party \
		--go_out=. \
		--go_opt=module=github.com/mibrgmv/whoami-server/history \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/mibrgmv/whoami-server/history \
		api/v1/*.proto
