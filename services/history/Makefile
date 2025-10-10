.PHONY: gen

gen:
	@mkdir -p internal/protogen/history
	protoc -I=api -I=../../shared/third_party \
		--go_out=internal/protogen/history \
		--go_opt=paths=source_relative \
		--go-grpc_out=internal/protogen/history \
		--go-grpc_opt=paths=source_relative \
		api/*.proto
