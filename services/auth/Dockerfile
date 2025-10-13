FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared ./shared

COPY services/auth ./services/auth

WORKDIR /app/services/auth
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o auth ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/auth/auth .
COPY --from=builder /app/services/auth/internal/config ./internal/config

EXPOSE 50055

CMD ["./auth"]