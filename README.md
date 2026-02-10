## архитектура бэкенда
![image](docs/whoami.png)
## как запустить
```shell
cd deployments/docker

# поднять core сервисы (gateway, auth, quiz, user, history, postgres, redis)
docker compose up -d

# поднять с keycloak
docker compose --profile keycloak up -d

# поднять с мониторингом (prometheus, grafana)
docker compose --profile monitoring up -d

# поднять всё
docker compose --profile keycloak --profile monitoring up -d

# запустить скрипт и получить секретный ключ
bash scripts/setup-keycloak.sh

# обновить значение `KEYCLOAK_ADMIN_CLIENT_SECRET` и пересоздать нужные сервисы
docker compose --profile keycloak up -d --force-recreate auth-service user-service
```
## `.env` для локального запуска
```dotenv
KEYCLOAK_BASE_URL=http://localhost:8088
KEYCLOAK_REALM=myrealm
KEYCLOAK_PUBLIC_CLIENT_ID=whoami-public
KEYCLOAK_PUBLIC_CLIENT_SECRET=
KEYCLOAK_ADMIN_CLIENT_ID=whoami-admin
KEYCLOAK_ADMIN_CLIENT_SECRET=<CHANGE_ME>
```

## возможные улучшения
- kafka
- исправить метрики
- накрутить nginx
- поднять несколько инстансов какого-то сервиса
- переделать главный сервис
- придумать темплейт для сервиса 
- отдавать фронт с бэка
- поднять пайплайн для гитхаба (тесты)
- добавить данные в миграции чтобы пользоваться из коробки
- переделать конфигурацию реалма
  - https://www.google.com/search?q=how+to+setup+keycloak+realm+on+startup+in+docker&sca_esv=50bdd2a08bdd7bce&ei=dedFaa5jr83A8A-ggfLgBQ&ved=0ahUKEwju8YS47sqRAxWvJhAIHaCAHFwQ4dUDCBA&uact=5&oq=how+to+setup+keycloak+realm+on+startup+in+docker&gs_lp=Egxnd3Mtd2l6LXNlcnAiMGhvdyB0byBzZXR1cCBrZXljbG9hayByZWFsbSBvbiBzdGFydHVwIGluIGRvY2tlcjIFEAAY7wUyCBAAGIAEGKIEMgUQABjvBTIIEAAYgAQYogQyCBAAGIAEGKIESLQjUO0HWKAgcAF4AZABAJgBogKgAfcTqgEGMC4xNi4xuAEDyAEA-AEBmAIKoAKpC8ICChAAGLADGNYEGEfCAgQQABgewgILEAAYgAQYhgMYigXCAggQIRigARjDBJgDAIgGAZAGCJIHBTEuNy4yoAevOrIHBTAuNy4yuAeiC8IHBTAuNi40yAcigAgA&sclient=gws-wiz-serp