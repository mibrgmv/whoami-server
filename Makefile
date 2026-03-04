COMPOSE := docker compose --profile monitoring

SERVICES := libs gateway quiz auth user history
BUILD_SERVICES := gateway quiz auth user history

.PHONY: up down down-v logs build lint test tidy gen help

up: build
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

down-v:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f

build:
	@for svc in $(BUILD_SERVICES); do \
		echo "==> Building $$svc"; \
		cd $$svc && make build && cd ..; \
	done

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

gen:
	@for svc in gateway auth quiz user history; do \
		echo "==> Generating proto for $$svc"; \
		cd $$svc && make gen && cd ..; \
	done

help:
	@echo "Docker:"
	@echo "  up      - build binaries + docker images and start"
	@echo "  down    - stop all"
	@echo "  down-v  - stop all + remove volumes"
	@echo "  logs    - follow logs"
	@echo ""
	@echo "Dev:"
	@echo "  build   - build all service binaries"
	@echo "  lint    - run golangci-lint"
	@echo "  test    - run tests"
	@echo "  tidy    - go mod tidy all modules"
	@echo "  gen     - generate proto files"
