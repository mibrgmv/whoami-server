FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

COPY configs/ ./configs/
COPY internal/ ./internal/
COPY protogen/ ./protogen/
COPY cmd/quizzes/ ./cmd/quizzes/

RUN go mod download && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o quizzes ./cmd/quizzes

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/quizzes .
COPY --from=builder /app/configs/ ./configs/

EXPOSE 50053

CMD ["./quizzes"]