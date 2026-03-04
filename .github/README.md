# CI/CD Pipeline

## Структура

```
.github/
├── actions/
│   └── ci-go/action.yml   # setup + lint + test
├── workflows/
│   ├── ci.yml    # lint + test
│   └── cd.yml    # build + deploy
└── README.md
```

## Триггеры

| Событие        | Ветки         | Workflow |
|----------------|---------------|----------|
| `push`         | `main`, `dev` | CI       |
| `pull_request` | `main`, `dev` | CI       |
| `push`         | `main`        | CD       |

## Что запускается (CI)

| Изменился             | Запустится     |
|-----------------------|----------------|
| `services/gateway/**` | только gateway |
| `services/quiz/**`    | только quiz    |
| `shared/**`           | ВСЕ сервисы    |
| `.golangci.yml`       | ВСЕ сервисы    |

## Jobs (CI)

```
changes ──► shared
        ├─► gateway
        ├─► quiz
        ├─► auth
        ├─► user
        └─► history
```

Каждый job вызывает `ci-go` action:
1. setup-go
2. go mod download
3. golangci-lint
4. go test

## Jobs (CD)

```
build-and-push ──► deploy
```

- `build-and-push` — собирает Docker образы для всех сервисов, пушит в ghcr.io
- `deploy` — SSH на сервер, docker compose pull && up

## Добавление нового сервиса

1. Добавить фильтр в `ci.yml`:
```yaml
# jobs.changes.steps.filter.with.filters
newservice:
  - 'services/newservice/**'
```

2. Добавить output:
```yaml
# jobs.changes.outputs
newservice: ${{ steps.filter.outputs.newservice }}
```

3. Добавить job:
```yaml
newservice:
  needs: changes
  if: needs.changes.outputs.newservice == 'true' || needs.changes.outputs.shared == 'true'
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: ./.github/actions/ci-go
      with:
        working-directory: services/newservice
```

## Локальный запуск

```bash
# Lint
cd services/quiz && golangci-lint run

# Tests
cd services/quiz && go test -race ./...

# Docker build
docker build --target quiz --build-arg SERVICE=quiz -t quiz:test .
```
