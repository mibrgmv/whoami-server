FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared ./shared

COPY services/history ./services/history

WORKDIR /app/services/history
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o history ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/history/history .
COPY --from=builder /app/services/history/internal/config ./internal/config
COPY --from=builder /app/services/history/internal/migrations ./internal/migrations

EXPOSE 50053

CMD ["./history"]