FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY internal/ ./internal/
COPY protogen/golang/user/ ./protogen/golang/user/
COPY protogen/golang/auth/ ./protogen/golang/auth/
COPY cmd/users/ ./cmd/users/

RUN go mod download && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o users ./cmd/users

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/users .
COPY --from=builder /app/cmd/users/configs/ ./cmd/users/configs/
COPY --from=builder /app/cmd/users/migrations/ ./cmd/users/migrations/

EXPOSE 50052

CMD ["./users"]