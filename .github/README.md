# CI/CD Pipeline

## Структура

```
.github/
├── actions/
│   ├── setup-go/action.yml   # setup Go + cache
│   ├── lint-go/action.yml    # golangci-lint
│   └── test-go/action.yml    # go test + coverage
├── workflows/
│   └── ci.yml                # основной пайплайн
└── README.md
```

## Триггеры

| Событие        | Ветки         |
|----------------|---------------|
| `push`         | `main`, `dev` |
| `pull_request` | `main`, `dev` |

## Что запускается

| Изменился             | Запустится     |
|-----------------------|----------------|
| `services/gateway/**` | только gateway |
| `services/quiz/**`    | только quiz    |
| `shared/**`           | ВСЕ сервисы    |
| `.github/**`          | ВСЕ сервисы    |
| `.golangci.yml`       | ВСЕ сервисы    |

## Jobs

```
changes ──► shared   ──►
        ├─► gateway  ──►
        ├─► quiz     ──► build (Docker)
        ├─► auth     ──►
        ├─► user     ──►
        └─► history  ──►
```

Каждый job использует actions:
- `lint-go` — golangci-lint
- `test-go` — go test с coverage

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
    - uses: ./.github/actions/lint-go
      with:
        module: services/newservice
    - uses: ./.github/actions/test-go
      with:
        module: services/newservice
```

4. Добавить в build matrix:
```yaml
- service: newservice
  run: ${{ needs.changes.outputs.newservice == 'true' || needs.changes.outputs.shared == 'true' }}
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
