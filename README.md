## архитектура бэкенда
![image](/whoami.png)
## как запустить
```shell
# поднять окружение
docker compose up -d

# запустить скрипт и получить секретный ключ
bash setup-keycloak.sh

# обновить значение `KEYCLOAK_ADMIN_CLIENT_SECRET` и пересоздать нужные сервисы
docker compose up -d --force-recreate gateway user-service
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