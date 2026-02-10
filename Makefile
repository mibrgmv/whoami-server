COMPOSE_DIR := deployments/docker
COMPOSE := docker compose -f $(COMPOSE_DIR)/docker-compose.yaml

SERVICES := shared gateway quiz auth user history

.PHONY: up down down-v logs \
	up-keycloak up-monitoring up-all \
	build lint test \
	lint-all test-all \
	proto tidy

up:
	$(COMPOSE) up -d

up-keycloak:
	$(COMPOSE) --profile keycloak up -d

up-monitoring:
	$(COMPOSE) --profile monitoring up -d

up-all:
	$(COMPOSE) --profile keycloak --profile monitoring up -d

down:
	$(COMPOSE) --profile keycloak --profile monitoring down

down-v:
	$(COMPOSE) --profile keycloak --profile monitoring down -v

logs:
	$(COMPOSE) logs -f

build:
	$(COMPOSE) build

lint:
	@for svc in $(SERVICES); do \
		echo "==> Linting $$svc"; \
		if [ "$$svc" = "shared" ]; then \
			cd $$svc && golangci-lint run --timeout=5m && cd ..; \
		else \
			cd services/$$svc && golangci-lint run --timeout=5m && cd ../..; \
		fi \
	done

test:
	@for svc in $(SERVICES); do \
		echo "==> Testing $$svc"; \
		if [ "$$svc" = "shared" ]; then \
			cd $$svc && go test -race ./... && cd ..; \
		else \
			cd services/$$svc && go test -race ./... && cd ../..; \
		fi \
	done

tidy:
	@for svc in $(SERVICES); do \
		echo "==> Tidying $$svc"; \
		if [ "$$svc" = "shared" ]; then \
			cd $$svc && go mod tidy && cd ..; \
		else \
			cd services/$$svc && go mod tidy && cd ../..; \
		fi \
	done

proto:
	@echo "Generate proto files (implement as needed)"

help:
	@echo "Docker:"
	@echo "  up            - start core services"
	@echo "  up-keycloak   - start with keycloak"
	@echo "  up-monitoring - start with prometheus/grafana"
	@echo "  up-all        - start everything"
	@echo "  down          - stop all"
	@echo "  down-v        - stop all + remove volumes"
	@echo "  logs          - follow logs"
	@echo "  build         - rebuild images"
	@echo ""
	@echo "Dev:"
	@echo "  lint          - run golangci-lint"
	@echo "  test          - run tests"
	@echo "  tidy          - go mod tidy all modules"
