FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared ./shared

COPY services/quizzes ./services/quizzes

WORKDIR /app/services/quizzes
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o quizzes ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/quizzes/quizzes .
COPY --from=builder /app/services/quizzes/internal/config ./internal/config
COPY --from=builder /app/services/quizzes/internal/migrations ./internal/migrations

EXPOSE 50051

CMD ["./quizzes"]