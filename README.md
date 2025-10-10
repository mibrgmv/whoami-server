запустить скрипт и получить секретный ключ
```shell
setup-keycloak.sh
```
обновить значение `KEYCLOAK_ADMIN_CLIENT_SECRET` и пересоздать гейтвей
```shell
docker compose up -d --force-recreate gateway
```