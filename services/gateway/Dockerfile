FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY services/gateway/go.mod services/gateway/go.sum ./services/gateway/
COPY shared/go.mod shared/go.sum ./shared/

WORKDIR /app/services/gateway
RUN go mod download

COPY shared/ ../../shared/
COPY services/gateway/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o gateway ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/gateway/gateway .
COPY --from=builder /app/services/gateway/internal/config ./internal/config
COPY --from=builder /app/services/gateway/api/v1/gateway.swagger.json ./api/v1/

EXPOSE 8080

CMD ["./gateway"]