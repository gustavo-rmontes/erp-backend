# Variáveis
PROJECT_NAME=erp-inteligente
DOCKER_COMPOSE=docker-compose
ENV_FILE=.env

# Comandos principais
help:
	@echo ""
	@echo "Comandos disponíveis:"
	@echo "  make build           => Build dos containers"
	@echo "  make up              => Sobe os serviços com Docker"
	@echo "  make down            => Derruba os serviços"
	@echo "  make restart         => Reinicia os serviços"
	@echo "  make backend-test    => Roda testes do Go"
	@echo "  make ai-test         => Roda testes dos agentes de IA"
	@echo "  make frontend-test   => Roda testes do Frontend"
	@echo "  make logs            => Mostra logs dos containers"
	@echo "  make prune           => Remove containers parados e volumes"
	@echo ""

# Build e execução
build:
	$(DOCKER_COMPOSE) build

up:
	$(DOCKER_COMPOSE) --env-file $(ENV_FILE) up -d

down:
	$(DOCKER_COMPOSE) down

restart: down up

logs:
	$(DOCKER_COMPOSE) logs -f --tail=100

prune:
	docker system prune -f
	docker volume prune -f

# Testes
backend-test:
	docker exec -it ${PROJECT_NAME}_backend go test ./...

ai-test:
	docker exec -it ${PROJECT_NAME}_ai pytest

frontend-test:
	docker exec -it ${PROJECT_NAME}_frontend npm run test
