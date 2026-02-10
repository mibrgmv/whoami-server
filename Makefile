.PHONY: \
	up-keycloak down-keycloak init-keycloak \
	up-monitoring down-monitoring \
	up down down-v


up-keycloak:
	cd deployments/docker && \
 	docker-compose -f docker-compose.keycloak.yaml up -d

down-keycloak:
	cd deployments/docker && docker-compose -f docker-compose.keycloak.yaml down

init-keycloak:
	bash scripts/setup-keycloak.sh


up-monitoring:
	cd deployments/docker && \
	docker-compose -f docker-compose.monitoring.yaml up -d

down-monitoring:
	cd deployments/docker && \
	docker-compose -f docker-compose.monitoring.yaml down


up:
	cd deployments/docker && \
 	docker-compose up -d

down:
	cd deployments/docker && docker-compose -f docker-compose.keycloak.yaml down
	cd deployments/docker && docker-compose -f docker-compose.monitoring.yaml down
	cd deployments/docker && docker-compose down

down-v:
	cd deployments/docker && docker-compose -f docker-compose.keycloak.yaml down -v
	cd deployments/docker && docker-compose -f docker-compose.monitoring.yaml down -v
	cd deployments/docker && docker-compose down -v