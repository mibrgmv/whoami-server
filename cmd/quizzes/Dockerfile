FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY internal/ ./internal/
COPY protogen/golang/history/ ./protogen/golang/history/
COPY protogen/golang/question/ ./protogen/golang/question/
COPY protogen/golang/quiz/ ./protogen/golang/quiz/
COPY cmd/quizzes/ ./cmd/quizzes/

RUN go mod download && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o quizzes ./cmd/quizzes

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/quizzes .
COPY --from=builder /app/cmd/quizzes/configs/ ./cmd/quizzes/configs/
COPY --from=builder /app/cmd/quizzes/migrations/ ./cmd/quizzes/migrations/

EXPOSE 50053

CMD ["./quizzes"]