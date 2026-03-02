COMPOSE := docker compose

SERVICES := libs gateway quiz auth user history

.PHONY: up down down-v logs \
	up-keycloak up-monitoring up-all \
	build lint test tidy proto help

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
		cd $$svc && golangci-lint run --timeout=5m && cd ..; \
	done

test:
	@for svc in $(SERVICES); do \
		echo "==> Testing $$svc"; \
		cd $$svc && go test -race ./... && cd ..; \
	done

tidy:
	@for svc in $(SERVICES); do \
		echo "==> Tidying $$svc"; \
		cd $$svc && go mod tidy && cd ..; \
	done

proto:
	@for svc in gateway auth quiz user history; do \
		echo "==> Generating proto for $$svc"; \
		cd $$svc && make gen && cd ..; \
	done

help:
	@echo "Docker:"
	@echo "  up            - start core services"
	@echo "  up-keycloak   - start with keycloak"
	@echo "  up-monitoring - start with prometheus"
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
	@echo "  proto         - generate proto files"
