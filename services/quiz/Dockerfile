FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared ./shared

COPY services/quiz ./services/quiz

WORKDIR /app/services/quiz
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o quiz ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/quiz/quiz .
COPY --from=builder /app/services/quiz/internal/config ./internal/config
COPY --from=builder /app/services/quiz/internal/migrations ./internal/migrations

EXPOSE 50051

CMD ["./quiz"]