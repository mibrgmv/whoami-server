FROM golang:1.25-alpine AS go-builder

ARG SERVICE

WORKDIR /src

COPY libs/go.mod libs/go.sum ./libs/
COPY ${SERVICE}/go.mod ${SERVICE}/go.sum ./${SERVICE}/

RUN go work init ./libs ./${SERVICE} && go mod download

COPY libs/ ./libs/
COPY ${SERVICE}/ ./${SERVICE}/

RUN CGO_ENABLED=0 go build -o /out/app ./${SERVICE}/cmd/app/


FROM alpine:latest AS runtime

RUN apk --no-cache add ca-certificates
WORKDIR /app


FROM runtime AS game

COPY --from=go-builder /out/app ./app
COPY game/config.yaml ./config.yaml
COPY game/migrations ./migrations

EXPOSE 50051
CMD ["./app"]


FROM runtime AS gateway

COPY --from=go-builder /out/app ./app
COPY gateway/config.yaml ./config.yaml
COPY gateway/api/v1/gateway.swagger.json ./api/v1/

EXPOSE 8080
CMD ["./app"]


FROM runtime AS statistics

COPY --from=go-builder /out/app ./app
COPY statistics/config.yaml ./config.yaml
COPY statistics/migrations ./migrations

EXPOSE 50053
CMD ["./app"]


FROM node:22-alpine AS web-builder

ARG VITE_API_BASE

WORKDIR /src
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN VITE_API_BASE=${VITE_API_BASE} npm run build


FROM nginx:alpine AS web

COPY --from=web-builder /src/dist /usr/share/nginx/html
COPY web/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
