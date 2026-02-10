FROM golang:1.24-alpine AS builder

ARG SERVICE
ARG USE_LOCAL_SHARED=false

WORKDIR /app

COPY shared/go.mod shared/go.sum ./shared/
COPY services/${SERVICE}/go.mod services/${SERVICE}/go.sum ./services/${SERVICE}/

WORKDIR /app/services/${SERVICE}

RUN if [ "$USE_LOCAL_SHARED" = "true" ]; then \
    go mod edit -replace github.com/mibrgmv/whoami-server/shared=../../shared; \
    fi

RUN go mod download

WORKDIR /app

COPY shared ./shared
COPY services/${SERVICE} ./services/${SERVICE}

WORKDIR /app/services/${SERVICE}

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/service ./cmd/*/


FROM alpine:latest AS runtime

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/service .


FROM runtime AS quiz

COPY --from=builder /app/services/quiz/internal/config ./internal/config
COPY --from=builder /app/services/quiz/internal/migrations ./internal/migrations

EXPOSE 50051

CMD ["./service"]


FROM runtime AS gateway

COPY --from=builder /app/services/gateway/internal/config ./internal/config
COPY --from=builder /app/services/gateway/api/v1/gateway.swagger.json ./api/v1/

EXPOSE 8080

CMD ["./service"]


FROM runtime AS auth

COPY --from=builder /app/services/auth/internal/config ./internal/config

EXPOSE 50055

CMD ["./service"]


FROM runtime AS user

COPY --from=builder /app/services/user/internal/config ./internal/config

EXPOSE 50052

CMD ["./service"]


FROM runtime AS history

COPY --from=builder /app/services/history/internal/config ./internal/config
COPY --from=builder /app/services/history/internal/migrations ./internal/migrations

EXPOSE 50053

CMD ["./service"]
