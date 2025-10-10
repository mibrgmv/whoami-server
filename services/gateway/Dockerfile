FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared ./shared

COPY services/gateway ./services/gateway

WORKDIR /app/services/gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/gateway/gateway .
COPY --from=builder /app/services/gateway/internal/config ./internal/config
COPY --from=builder /app/services/gateway/api/swagger ./api/swagger

EXPOSE 8080

CMD ["./gateway"]