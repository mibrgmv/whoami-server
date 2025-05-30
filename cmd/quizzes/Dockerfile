FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o whoami ./cmd/whoami

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/whoami .

EXPOSE 50053

CMD ["./whoami"]