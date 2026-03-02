FROM golang:1.25-alpine AS builder

ARG SERVICE

WORKDIR /app

COPY go.work ./
COPY libs ./libs
COPY ${SERVICE} ./${SERVICE}

WORKDIR /app/${SERVICE}

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/service ./cmd/*/


FROM alpine:latest AS runtime

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/service .


FROM runtime AS quiz

COPY --from=builder /app/quiz/config.yaml ./config.yaml
COPY --from=builder /app/quiz/migrations ./migrations

EXPOSE 50051

CMD ["./service"]


FROM runtime AS gateway

COPY --from=builder /app/gateway/config.yaml ./config.yaml
COPY --from=builder /app/gateway/api/v1/gateway.swagger.json ./api/v1/

EXPOSE 8080

CMD ["./service"]


FROM runtime AS auth

COPY --from=builder /app/auth/config.yaml ./config.yaml

EXPOSE 50055

CMD ["./service"]


FROM runtime AS user

COPY --from=builder /app/user/config.yaml ./config.yaml

EXPOSE 50052

CMD ["./service"]


FROM runtime AS history

COPY --from=builder /app/history/config.yaml ./config.yaml
COPY --from=builder /app/history/migrations ./migrations

EXPOSE 50053

CMD ["./service"]
