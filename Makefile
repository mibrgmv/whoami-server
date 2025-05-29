include .env

migrate_up:
	migrate -path=/migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=prefer" -verbose up

migrate_down:
	migrate -path=/migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=prefer" -verbose down

lint:
	golangci-lint run

swagger:
	swag init -o ./docs -d ./cmd

protoc:
	protoc ./cmd/*/api/*.proto --go_out=.. --go-grpc_out=..