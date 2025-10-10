FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared ./shared

COPY services/users ./services/users

WORKDIR /app/services/users
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o users ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/users/users .
COPY --from=builder /app/services/users/internal/config ./internal/config

EXPOSE 50052

CMD ["./users"]