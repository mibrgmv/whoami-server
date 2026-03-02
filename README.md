## структура проекта

```
/
├── auth/           # auth сервис (gRPC)
├── gateway/        # HTTP gateway
├── history/        # history сервис
├── quiz/           # quiz сервис (legacy)
├── user/           # user сервис
├── libs/           # shared библиотеки
├── docker-compose.yaml
├── docker-compose.override.yaml
├── Dockerfile
├── go.work
├── realm.json      # keycloak realm config
└── prometheus.yml  # prometheus config
```

## запуск

```shell
# core сервисы (локальная сборка)
docker compose up -d

# + keycloak
docker compose --profile keycloak up -d

# + prometheus
docker compose --profile monitoring up -d

# всё
docker compose --profile keycloak --profile monitoring up -d
```

## разработка

```shell
make lint    # golangci-lint
make test    # тесты
make tidy    # go mod tidy
make proto   # генерация proto
```

## keycloak

Realm импортируется автоматически из `realm.json`.

Тестовые юзеры:
- `admin:admin` — роли: user, admin
