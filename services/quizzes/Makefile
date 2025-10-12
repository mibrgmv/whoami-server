.PHONY: lint gen


lint:
	golangci-lint run

gen:
	protoc -I=api/v1 -I=../../shared/third_party \
		--go_out=. \
		--go_opt=module=github.com/mibrgmv/whoami-server/quizzes \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/mibrgmv/whoami-server/quizzes \
		api/v1/*.proto