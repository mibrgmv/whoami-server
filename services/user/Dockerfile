FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared ./shared

COPY services/user ./services/user

WORKDIR /app/services/user
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o user ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/user/user .
COPY --from=builder /app/services/user/internal/config ./internal/config

EXPOSE 50052

CMD ["./user"]