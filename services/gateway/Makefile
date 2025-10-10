.PHONY: lint gen swag build push

PROTO_DIR=api
PROTOGEN_DIR=internal/protogen
PROTO_FILES=$(wildcard $(PROTO_DIR)/*.proto)
THIRD_PARTY_DIR=../../shared/third_party
SERVICES = $(basename $(notdir $(PROTO_FILES)))

lint:
	golangci-lint run

gen:
	@mkdir -p $(PROTOGEN_DIR)
	$(foreach service,$(SERVICES),\
		mkdir -p $(PROTOGEN_DIR)/$(service); \
		protoc -I=$(PROTO_DIR) -I=$(THIRD_PARTY_DIR) \
			--go_out=$(PROTOGEN_DIR)/$(service) \
			--go_opt=paths=source_relative \
			--go-grpc_out=$(PROTOGEN_DIR)/$(service) \
			--go-grpc_opt=paths=source_relative \
			--grpc-gateway_out=$(PROTOGEN_DIR)/$(service) \
			--grpc-gateway_opt=paths=source_relative \
			$(PROTO_DIR)/$(service).proto; \
	)
	protoc -I=$(PROTO_DIR) -I=$(THIRD_PARTY_DIR) \
		--openapiv2_out=$(PROTO_DIR) \
		--openapiv2_opt=logstderr=true \
		$(PROTO_DIR)/*.proto

swag:
	swag init -g cmd/app/main.go -o api/swagger/custom --parseDependency --parseInternal
