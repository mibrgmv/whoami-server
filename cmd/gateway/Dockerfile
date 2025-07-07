FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY internal/ ./internal/
COPY protogen/ ./protogen/
COPY cmd/gateway/ ./cmd/gateway/

RUN go mod download && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./cmd/gateway

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gateway .
COPY --from=builder /app/cmd/gateway/configs/ ./cmd/gateway/configs/

EXPOSE 8080

CMD ["./gateway"]