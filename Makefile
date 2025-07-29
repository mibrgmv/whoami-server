lint:
	golangci-lint run

gen:
	/bin/bash gen.sh

swag:
	swag init -g cmd/app/main.go -o api/swagger --parseDependency --parseInternal

.PHONY: lint gen swag