# Guia de Testes - Sistema de Leilões

Este documento explica como executar os testes do projeto usando diferentes métodos.

## 🚀 Métodos de Execução de Testes

### 1. Testes Locais (Desenvolvimento)

#### Pré-requisitos
- Go 1.20+
- MongoDB rodando localmente

#### Comandos
```bash
# Executar todos os testes
go test ./... -v -cover

# Executar testes específicos do auction
go test ./internal/infra/database/auction/ -v -cover

# Executar testes com cobertura
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 2. Testes com Docker Compose

#### Opção A: Makefile

```bash
# Ver todos os comandos disponíveis
make help

# Executar todos os testes
make docker-test

# Executar testes do auction
make docker-test-auction

# Executar testes de integração
make docker-test-integration

# Limpar ambiente
make docker-clean
```

#### Opção B: Comandos Docker Compose Diretos

```bash
# Executar todos os testes
docker-compose -f docker-compose.test.yml up --build test-all

# Executar testes do auction
docker-compose -f docker-compose.test.yml up --build test-auction

# Executar testes de integração
docker-compose -f docker-compose.test.yml up --build test-integration

# Limpar ambiente
docker-compose -f docker-compose.test.yml down -v --remove-orphans
```

## 📊 Tipos de Testes

### 1. Testes Unitários
- **Localização**: `internal/infra/database/auction/auction_test.go`
- **Foco**: Funcionalidades específicas do fechamento automático
- **Execução**: `go test ./internal/infra/database/auction/ -v`

### 2. Testes de Integração
- **Foco**: Interação entre componentes
- **Execução**: `go test ./... -v -tags=integration`

### 3. Testes de Cobertura
- **Objetivo**: Verificar cobertura de código
- **Execução**: `go test ./... -cover -coverprofile=coverage.out`

## 🔧 Configurações de Teste

### Variáveis de Ambiente para Testes

```env
# Configurações otimizadas para testes
MONGODB_URI=mongodb://mongodb-test:27017
MONGODB_DATABASE=test_auction_db
AUCTION_INTERVAL=30s
BATCH_INSERT_INTERVAL=3s
MAX_BATCH_SIZE=4
GIN_MODE=test
```

### Estrutura de Testes

```
internal/infra/database/auction/
├── auction_test.go          # Testes unitários
├── create_auction.go        # Implementação
└── find_auction.go         # Implementação
```

## 🐳 Docker Compose para Testes

### Arquivo: `docker-compose.test.yml`

```yaml
services:
  mongodb-test:
    image: mongo:latest
    ports:
      - "27018:27017"
    environment:
      - MONGO_INITDB_DATABASE=test_auction_db

  test-all:
    build:
      dockerfile: Dockerfile
      context: .
    command: sh -c "go test ./... -v -cover"
    depends_on:
      - mongodb-test

  test-auction:
    build:
      dockerfile: Dockerfile
      context: .
    command: sh -c "go test ./internal/infra/database/auction/ -v -cover"
    depends_on:
      - mongodb-test
```

## 📈 Relatórios de Cobertura

### Gerar Relatório de Cobertura

```bash
# Gerar arquivo de cobertura
go test ./... -cover -coverprofile=coverage.out

# Visualizar relatório HTML
go tool cover -html=coverage.out -o coverage.html

# Visualizar relatório no terminal
go tool cover -func=coverage.out
```

### Cobertura Esperada

- **Funções de configuração**: 100%
- **Goroutine de fechamento**: 100%
- **Métodos de atualização**: 100%
- **Controle de concorrência**: 100%

## 🚨 Solução de Problemas

### Problema: MongoDB não disponível
```bash
# Solução: Usar Docker Compose
docker-compose -f docker-compose.test.yml up mongodb-test
```

### Problema: Testes falhando por timeout
```bash
# Solução: Aumentar o intervalo do leilão nos testes
export AUCTION_INTERVAL=60s
```

### Problema: Containers não param
```bash
# Solução: Limpar ambiente
docker-compose -f docker-compose.test.yml down -v --remove-orphans
docker system prune -f
```

## 📋 Checklist de Testes

### Antes de Fazer Commit

- [ ] Testes unitários passando
- [ ] Testes de integração passando
- [ ] Cobertura de código > 80%
- [ ] Sem vazamentos de memória
- [ ] Logs de teste limpos

### Comandos de Verificação

```bash
# Verificar se todos os testes passam
make docker-test

# Verificar cobertura
go test ./... -cover

# Verificar se não há vazamentos
go test ./... -race
```

## 🎯 Exemplos de Uso

### Teste Rápido de Desenvolvimento
```bash
# Executar apenas testes do auction
make docker-test-auction
```

### Teste Completo antes do Deploy
```bash
# Executar todos os testes com cobertura
make docker-test
```

### Limpeza de Ambiente
```bash
# Limpar tudo e começar do zero
make docker-clean
```

## 📚 Recursos Adicionais

- [Documentação Go Testing](https://golang.org/pkg/testing/)
- [Docker Compose Testing](https://docs.docker.com/compose/)
- [Go Coverage](https://blog.golang.org/cover)
