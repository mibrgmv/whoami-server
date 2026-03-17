COMPOSE := docker compose

SERVICES := libs gateway game iam statistics
BUILD_SERVICES := gateway game iam statistics

.PHONY: up down down-v rs logs build build-all lint test tidy gen web web-build web-install help

up: build-all
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

down-v:
	$(COMPOSE) down -v

rs: down up

logs:
	$(COMPOSE) logs -f

build:
	@for svc in $(BUILD_SERVICES); do \
		echo "==> Building $$svc"; \
		cd $$svc && make build && cd ..; \
	done

build-all: build web-build

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
	@for svc in gateway iam game statistics; do \
		echo "==> Generating proto for $$svc"; \
		cd $$svc && make gen && cd ..; \
	done

web-install:
	cd web && npm install

web:
	cd web && npm run dev

web-build:
	cd web && npm run build

help:
	@echo "Docker:"
	@echo "  up      - build binaries + docker images and start"
	@echo "  down    - stop all"
	@echo "  down-v  - stop all + remove volumes"
	@echo "  rs      - restart all (down + up with rebuild)"
	@echo "  logs    - follow logs"
	@echo ""
	@echo "Build:"
	@echo "  build       - build backend service binaries"
	@echo "  build-all   - build backend + frontend"
	@echo "  web-build   - build frontend for production"
	@echo ""
	@echo "Dev:"
	@echo "  lint        - run golangci-lint"
	@echo "  test        - run tests"
	@echo "  tidy        - go mod tidy all modules"
	@echo "  gen         - generate proto files"
	@echo "  web         - run frontend dev server (port 3000)"
	@echo "  web-install - npm install for frontend"
