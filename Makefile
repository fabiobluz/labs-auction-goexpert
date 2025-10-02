# Makefile para o projeto de leilões
# Facilita a execução de comandos comuns

.PHONY: help build run test test-auction test-integration clean docker-build docker-run docker-test

# Variáveis
DOCKER_COMPOSE = docker-compose
DOCKER_COMPOSE_TEST = docker-compose -f docker-compose.test.yml

# Comando padrão
help: ## Mostra esta ajuda
	@echo "Comandos disponíveis:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Comandos de desenvolvimento
build: ## Compila o projeto
	go build ./cmd/auction

run: ## Executa o projeto localmente
	go run cmd/auction/main.go

# Comandos de teste
test: ## Executa todos os testes
	go test ./... -v -cover

test-auction: ## Executa testes específicos do auction
	go test ./internal/infra/database/auction/ -v -cover

test-integration: ## Executa testes de integração
	go test ./... -v -cover -tags=integration

# Comandos Docker
docker-build: ## Constrói a imagem Docker
	$(DOCKER_COMPOSE) build

docker-run: ## Executa o projeto com Docker Compose
	$(DOCKER_COMPOSE) up --build

docker-test: ## Executa testes com Docker Compose
	$(DOCKER_COMPOSE_TEST) up --build test-all

docker-test-auction: ## Executa testes do auction com Docker Compose
	$(DOCKER_COMPOSE_TEST) up --build test-auction

docker-test-integration: ## Executa testes de integração com Docker Compose
	$(DOCKER_COMPOSE_TEST) up --build test-integration

docker-test-auto-close: ## Executa testes de fechamento automático com Docker Compose
	$(DOCKER_COMPOSE_TEST) up --build test-auto-close

docker-test-performance: ## Executa testes de performance com Docker Compose
	$(DOCKER_COMPOSE_TEST) up --build test-performance

# Comandos de limpeza
clean: ## Limpa arquivos temporários
	go clean
	rm -f auction

docker-clean: ## Limpa containers e volumes Docker
	$(DOCKER_COMPOSE) down -v --remove-orphans
	$(DOCKER_COMPOSE_TEST) down -v --remove-orphans
	docker system prune -f

# Comandos de desenvolvimento completo
dev-setup: ## Configura ambiente de desenvolvimento
	@echo "Configurando ambiente de desenvolvimento..."
	@if [ ! -f cmd/auction/.env ]; then \
		cp env.example cmd/auction/.env; \
		echo "Arquivo .env criado a partir do env.example"; \
	fi
	@echo "Ambiente configurado!"

dev-test: ## Executa testes em ambiente de desenvolvimento
	@echo "Executando testes de desenvolvimento..."
	go test ./... -v -cover -short

# Comandos de produção
prod-build: ## Constrói para produção
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auction ./cmd/auction

prod-test: ## Executa testes de produção
	$(DOCKER_COMPOSE_TEST) up --build test-all
