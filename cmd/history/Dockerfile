FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY internal/ ./internal/
COPY protogen/golang/history/ ./protogen/golang/history/
COPY cmd/history/ ./cmd/history/

RUN go mod download && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o history ./cmd/history

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/history .
COPY --from=builder /app/cmd/history/configs/ ./cmd/history/configs/
COPY --from=builder /app/cmd/history/migrations/ ./cmd/history/migrations/

EXPOSE 50051

CMD ["./history"]