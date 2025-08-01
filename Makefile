REGISTRY = ghcr.io/your-username
SERVICE = whoami-server-gateway
TAG = latest
DOCKERFILE = ./cmd/app/Dockerfile

lint:
	golangci-lint run

gen:
	/bin/bash gen.sh

swag:
	swag init -g cmd/app/main.go -o api/swagger --parseDependency --parseInternal

build:
	docker build -t $(SERVICE):$(TAG) -f $(DOCKERFILE) .

push: build
	docker tag $(SERVICE):$(TAG) $(REGISTRY)/$(SERVICE):$(TAG)
	docker push $(REGISTRY)/$(SERVICE):$(TAG)

.PHONY: lint gen swag build push