# CI/CD Pipeline

## Обзор

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              CI Pipeline                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐               │
│  │ Lint/shared  │    │ Test/shared  │    │Build/gateway │               │
│  │ Lint/gateway │    │ Test/gateway │    │ Build/quiz   │               │
│  │ Lint/quiz    │───▶│ Test/quiz    │───▶│ Build/auth   │───▶Integration│
│  │ Lint/auth    │    │ Test/auth    │    │ Build/user   │               │
│  │ Lint/user    │    │ Test/user    │    │Build/history │               │
│  │ Lint/history │    │ Test/history │    └──────────────┘               │
│  └──────────────┘    └──────────────┘                                   │
│     (parallel)          (parallel)          (parallel)                  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

## Триггеры

| Событие | Ветки |
|---------|-------|
| `push` | `main`, `dev` |
| `pull_request` | `main` |

## Jobs

### 1. Lint

Запускает `golangci-lint` для каждого Go модуля параллельно.

- **Конфиг:** `/.golangci.yml`
- **Timeout:** 5 минут
- **Линтеры:** errcheck, govet, staticcheck, bodyclose, sqlclosecheck, и др.

### 2. Test

Запускает unit тесты с race detector и собирает coverage.

- **Флаги:** `-v -race -coverprofile=coverage.out -covermode=atomic`
- **Coverage:** отправляется в Codecov (нужен `CODECOV_TOKEN` в secrets)

### 3. Build

Собирает Docker образы для каждого сервиса.

- **Dockerfile:** `/Dockerfile` (multi-stage)
- **Cache:** GitHub Actions cache (`type=gha`)
- **Push:** отключен (только проверка сборки)

### 4. Integration

Поднимает зависимости и запускает интеграционные тесты.

- **Зависимости:** postgres-quiz, postgres-history, redis
- **Тесты:** `go test -tags=integration`

## Composite Actions

Переиспользуемые actions в `/.github/actions/`:

### setup-go

Setup Go окружения с кэшированием.

```yaml
- uses: ./.github/actions/setup-go
  with:
    module: services/quiz      # обязательный
    go-version: '1.24'         # опциональный, default: 1.24
```

### lint-go

Запуск golangci-lint.

```yaml
- uses: ./.github/actions/lint-go
  with:
    module: services/quiz      # обязательный
    go-version: '1.24'         # опциональный
```

### test-go

Запуск тестов с coverage.

```yaml
- uses: ./.github/actions/test-go
  with:
    module: services/quiz      # обязательный
    go-version: '1.24'         # опциональный
    upload-coverage: 'true'    # опциональный, default: true
    codecov-token: ${{ secrets.CODECOV_TOKEN }}  # опциональный
```

## Добавление нового сервиса

1. Добавить в matrix в `ci.yml`:

```yaml
matrix:
  module:
    - shared
    - services/gateway
    - services/quiz
    - services/auth
    - services/user
    - services/history
    - services/new-service  # добавить сюда
```

2. Для build job (если нужен Docker):

```yaml
matrix:
  service:
    - gateway
    - quiz
    - auth
    - user
    - history
    - new-service  # добавить сюда
```

## Secrets

| Secret | Описание | Обязательный |
|--------|----------|--------------|
| `CODECOV_TOKEN` | Токен для загрузки coverage | Нет (CI не упадёт) |

## Локальный запуск

### Lint
```bash
cd services/quiz
golangci-lint run
```

### Tests
```bash
cd services/quiz
go test -v -race ./...
```

### Integration tests
```bash
cd deployments/docker
docker compose up -d postgres-quiz redis
cd ../../services/quiz
go test -v -tags=integration ./...
```

## Структура файлов

```
.github/
├── actions/
│   ├── setup-go/
│   │   └── action.yml
│   ├── lint-go/
│   │   └── action.yml
│   └── test-go/
│       └── action.yml
├── workflows/
│   └── ci.yml
└── README.md              # этот файл
```
