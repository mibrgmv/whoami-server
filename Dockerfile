FROM alpine:latest AS runtime

RUN apk --no-cache add ca-certificates

WORKDIR /app


FROM runtime AS game

COPY game/bin/app ./app
COPY game/config.yaml ./config.yaml
COPY game/migrations ./migrations

EXPOSE 50051

CMD ["./app"]


FROM runtime AS gateway

COPY gateway/bin/app ./app
COPY gateway/config.yaml ./config.yaml
COPY gateway/api/v1/gateway.swagger.json ./api/v1/

EXPOSE 8080

CMD ["./app"]


FROM runtime AS iam

COPY iam/bin/app ./app
COPY iam/config.yaml ./config.yaml

EXPOSE 50052

CMD ["./app"]


FROM runtime AS statistics

COPY statistics/bin/app ./app
COPY statistics/config.yaml ./config.yaml
COPY statistics/migrations ./migrations

EXPOSE 50053

CMD ["./app"]


FROM nginx:alpine AS web

COPY web/dist /usr/share/nginx/html
COPY web/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
