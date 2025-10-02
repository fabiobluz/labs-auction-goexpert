# Testes Docker

Este documento descreve como executar testes usando Docker e Docker Compose.

## Pré-requisitos

- Docker Desktop instalado e rodando
- Docker Compose disponível
- Acesso à internet para baixar imagens

## Estrutura dos Testes

### Tipos de Testes

1. **Testes Unitários com Mock** (`auction_test.go`)
   - Não dependem de infraestrutura externa
   - Executam rapidamente
   - Usam mocks para simular MongoDB

2. **Testes de Integração com Docker** (`auction_docker_test.go`)
   - Dependem do MongoDB real
   - Testam fechamento automático real
   - Executam em ambiente isolado

### Serviços Docker

- **mongodb-test**: MongoDB para testes
- **test-all**: Executa todos os testes
- **test-auction**: Executa testes específicos do auction
- **test-integration**: Executa testes de integração
- **test-auto-close**: Executa testes de fechamento automático
- **test-performance**: Executa testes de performance

## Executando Testes

### Usando Makefile

```bash
# Todos os testes
make docker-test

# Testes específicos do auction
make docker-test-auction

# Testes de integração
make docker-test-integration

# Testes de fechamento automático
make docker-test-auto-close

# Testes de performance
make docker-test-performance
```

### Usando Docker Compose Diretamente

```bash
# Todos os testes
docker-compose -f docker-compose.test.yml up --build test-all

# Testes específicos
docker-compose -f docker-compose.test.yml up --build test-auction
docker-compose -f docker-compose.test.yml up --build test-auto-close
docker-compose -f docker-compose.test.yml up --build test-performance
```

### Usando Scripts

#### PowerShell (Windows)
```powershell
# Todos os testes
.\scripts\test-docker.ps1 all

# Testes específicos
.\scripts\test-docker.ps1 auction
.\scripts\test-docker.ps1 auto-close
.\scripts\test-docker.ps1 performance

# Com limpeza
.\scripts\test-docker.ps1 all -Clean
```

#### Bash (Linux/Mac)
```bash
# Todos os testes
./scripts/test-docker.sh all

# Testes específicos
./scripts/test-docker.sh auction
./scripts/test-docker.sh auto-close
./scripts/test-docker.sh performance

# Com limpeza
./scripts/test-docker.sh all --clean
```

## Configuração

### Variáveis de Ambiente

As variáveis de ambiente são configuradas no `docker-compose.test.yml`:

```yaml
environment:
  - MONGODB_URL=mongodb://mongodb-test:27017
  - MONGODB_DB=test_auction_db
  - AUCTION_INTERVAL=2s
  - AUCTION_CHECK_INTERVAL=1s
  - AUCTION_CONTEXT_TIMEOUT=10s
  - GIN_MODE=test
```

### Configuração de Testes

- **AUCTION_INTERVAL**: Intervalo de fechamento automático (padrão: 2s para testes)
- **AUCTION_CHECK_INTERVAL**: Intervalo de verificação (padrão: 1s para testes)
- **AUCTION_CONTEXT_TIMEOUT**: Timeout para operações (padrão: 10s para testes)

## Tipos de Testes

### 1. Testes de Fechamento Automático

Testam o fechamento automático de leilões usando MongoDB real:

```go
func TestAutoCloseAuctionWithMongoDB(t *testing.T) {
    // Configura intervalo curto para teste
    os.Setenv("AUCTION_INTERVAL", "2s")
    
    // Cria leilão
    // Aguarda fechamento automático
    // Verifica status
}
```

### 2. Testes de Performance

Testam a performance com múltiplos leilões:

```go
func TestPerformanceWithAutoClose(t *testing.T) {
    // Cria 10 leilões
    // Mede tempo de criação
    // Verifica fechamento automático
}
```

### 3. Testes de Concorrência

Testam operações simultâneas:

```go
func TestConcurrentAuctionCreationWithMongoDB(t *testing.T) {
    // Cria múltiplos leilões rapidamente
    // Verifica fechamento automático
}
```

## Limpeza

### Limpar Containers e Volumes

```bash
# Usando Makefile
make docker-clean

# Usando Docker Compose
docker-compose -f docker-compose.test.yml down -v --remove-orphans
docker system prune -f
```

### Limpar Apenas Volumes de Teste

```bash
docker volume rm labs-auction-goexpert_mongo-test-data
```

## Troubleshooting

### MongoDB Não Conecta

1. Verifique se o Docker está rodando
2. Verifique se a porta 27018 está disponível
3. Aguarde o healthcheck do MongoDB (30s)

### Testes Falham por Timeout

1. Aumente o `AUCTION_INTERVAL` nos testes
2. Verifique se o MongoDB está saudável
3. Verifique logs: `docker-compose -f docker-compose.test.yml logs`

### Problemas de Rede

1. Verifique se a rede `testNetwork` está criada
2. Reinicie o Docker Desktop
3. Limpe containers órfãos: `docker system prune -f`

## Logs e Debug

### Ver Logs dos Testes

```bash
docker-compose -f docker-compose.test.yml logs test-auction
```

### Ver Logs do MongoDB

```bash
docker-compose -f docker-compose.test.yml logs mongodb-test
```

### Executar Container Interativo

```bash
docker-compose -f docker-compose.test.yml run --rm test-auction sh
```

## Exemplos de Uso

### Executar Testes Específicos

```bash
# Apenas testes de fechamento automático
make docker-test-auto-close

# Apenas testes de performance
make docker-test-performance
```

### Executar com Configuração Customizada

```bash
# Modificar docker-compose.test.yml para ajustar variáveis
# Executar testes
docker-compose -f docker-compose.test.yml up --build test-auto-close
```

### Executar Testes em Paralelo

```bash
# Executar múltiplos tipos de teste
docker-compose -f docker-compose.test.yml up --build test-auction test-performance
```

## Monitoramento

### Verificar Status dos Containers

```bash
docker-compose -f docker-compose.test.yml ps
```

### Verificar Uso de Recursos

```bash
docker stats
```

### Verificar Logs em Tempo Real

```bash
docker-compose -f docker-compose.test.yml logs -f test-auction
```
