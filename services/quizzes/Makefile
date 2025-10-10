.PHONY: gen

PROTO_DIR=api
PROTOGEN_DIR=internal/protogen
PROTO_FILES=$(wildcard $(PROTO_DIR)/*.proto)
THIRD_PARTY_DIR=../../shared/third_party
SERVICES = $(basename $(notdir $(PROTO_FILES)))

gen:
	@mkdir -p $(PROTOGEN_DIR)
	$(foreach service,$(SERVICES),\
		mkdir -p $(PROTOGEN_DIR)/$(service); \
		protoc -I=$(PROTO_DIR) -I=$(THIRD_PARTY_DIR) \
			--go_out=$(PROTOGEN_DIR)/$(service) \
			--go_opt=paths=source_relative \
			--go-grpc_out=$(PROTOGEN_DIR)/$(service) \
			--go-grpc_opt=paths=source_relative \
			$(PROTO_DIR)/$(service).proto; \
	)
