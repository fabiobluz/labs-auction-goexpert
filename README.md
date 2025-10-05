# Sistema de Leilões com Fechamento Automático

Este projeto implementa um sistema de leilões em Go com funcionalidade de fechamento automático baseado em tempo configurável.

## Funcionalidades

- ✅ Criação de leilões
- ✅ Sistema de lances (bids)
- ✅ Fechamento automático de leilões baseado em tempo
- ✅ Validação de leilões fechados para novos lances
- ✅ API REST para gerenciamento

## 📚 Documentação

- [Guia de Testes](docs/TESTING.md) - Como executar testes locais e com Docker
- [Testes Docker](docs/DOCKER_TESTING.md) - Detalhes sobre testes com Docker Compose
- [Resumo da Implementação](docs/IMPLEMENTATION_SUMMARY.md) - Detalhes técnicos da implementação

## Arquitetura

O projeto segue os princípios de Clean Architecture:

```
cmd/auction/                    # Ponto de entrada da aplicação
internal/
├── entity/                     # Entidades de domínio
├── usecase/                    # Casos de uso
├── infra/
│   ├── api/web/controller/     # Controladores HTTP
│   └── database/               # Repositórios e acesso a dados
configuration/                  # Configurações (logger, database, etc.)
```

## Nova Funcionalidade: Fechamento Automático

### Implementação

A funcionalidade de fechamento automático foi implementada no arquivo `internal/infra/database/auction/create_auction.go` com as seguintes características:

1. **Timer Automático**: Usa goroutine com timer para fechar leilões após o tempo configurado
2. **Controle de Tempo**: Utiliza variáveis de ambiente para configurar o tempo de duração dos leilões
3. **Inserção em Lote**: Sistema de batch insert para lances, otimizando performance
4. **Logging**: Registra automaticamente quando leilões são fechados

### Variáveis de Ambiente

- `AUCTION_INTERVAL`: Duração do leilão (ex: "2m", "30s", "1h")
- `MONGODB_URI`: URI de conexão com MongoDB
- `MONGODB_DATABASE`: Nome do banco de dados
- `BATCH_INSERT_INTERVAL`: Intervalo para inserção de lances em lote (ex: "3s", "5s", "1m")
- `MAX_BATCH_SIZE`: Tamanho máximo do lote de lances (ex: 5, 10, 20)

## Como Executar

### Pré-requisitos

- Docker
- Docker Compose
- Go 1.20+ (para desenvolvimento local)

### Execução com Docker

1. **Clone o repositório**:
```bash
git clone <repository-url>
cd labs-auction-goexpert
```

2. **Execute com Docker Compose**:
```bash
docker-compose up --build
```

3. **Acesse a aplicação**:
- API: http://localhost:8080
- MongoDB: localhost:27017

### Execução Local (Desenvolvimento)

1. **Instale as dependências**:
```bash
go mod download
```

2. **Configure as variáveis de ambiente**:
```bash
export MONGODB_URI=mongodb://localhost:27017
export MONGODB_DATABASE=auction_db
export AUCTION_INTERVAL=2m
export BATCH_INSERT_INTERVAL=3m
export MAX_BATCH_SIZE=5
export GIN_MODE=debug
```

3. **Execute o MongoDB**:
```bash
docker run -d -p 27017:27017 mongo:latest
```

4. **Execute a aplicação**:
```bash
go run cmd/auction/main.go
```

## Testes

### Executar Testes Unitários

```bash
# Testes gerais
go test ./...

# Testes específicos do fechamento automático
go test ./internal/infra/database/auction/ -v

# Testes com cobertura
go test ./... -cover
```

### Teste Manual do Fechamento Automático

1. **Crie um leilão**:
```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15",
    "category": "Electronics",
    "description": "Latest iPhone model",
    "condition": 1
  }'
```

2. **Aguarde o tempo configurado** (padrão: 2 minutos)

3. **Verifique se o leilão foi fechado**:
```bash
curl http://localhost:8080/auction/{auction_id}
```

## API Endpoints

### Leilões
- `GET /auction` - Lista leilões
- `GET /auction/:auctionId` - Busca leilão por ID
- `POST /auction` - Cria novo leilão
- `GET /auction/winner/:auctionId` - Busca lance vencedor

### Lances
- `POST /bid` - Cria novo lance
- `GET /bid/:auctionId` - Lista lances de um leilão

### Usuários
- `GET /user/:userId` - Busca usuário por ID

## Configuração Avançada

### Configurações de Ambiente

Crie um arquivo `.env` na pasta `cmd/auction/`:

```env
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DATABASE=auction_db
AUCTION_INTERVAL=2m
BATCH_INSERT_INTERVAL=3m
MAX_BATCH_SIZE=5
GIN_MODE=debug
```

## Monitoramento

### Logs de Fechamento Automático

O sistema registra automaticamente quando leilões são fechados:

```
INFO: Auction closed automatically {"auctionId": "uuid", "timestamp": "2024-01-01T12:00:00Z"}
```

### Verificação de Status

Para verificar se a goroutine está funcionando, monitore os logs da aplicação. Você deve ver logs de verificação a cada 10 segundos.

## Solução de Problemas

### Leilões não estão fechando automaticamente

1. Verifique se a variável `AUCTION_INTERVAL` está configurada
2. Confirme se o MongoDB está acessível
3. Verifique os logs da aplicação

### Erro de conexão com MongoDB

1. Verifique se o MongoDB está rodando
2. Confirme a URI de conexão
3. Verifique as permissões de rede

### Performance

- O sistema usa timers individuais para cada leilão
- Lances são inseridos em lote para otimizar performance
- Ajuste `BATCH_INSERT_INTERVAL` e `MAX_BATCH_SIZE` conforme necessário

## Estrutura do Código

### Principais Arquivos Modificados

- `internal/infra/database/auction/create_auction.go`: Implementação do fechamento automático
- `internal/entity/auction_entity/auction_entity.go`: Interface atualizada
- `docker-compose.yml`: Configuração de ambiente
- `internal/infra/database/auction/auction_test.go`: Testes unitários

### Padrões Utilizados

- **Repository Pattern**: Para acesso a dados
- **Use Case Pattern**: Para lógica de negócio
- **Dependency Injection**: Para injeção de dependências
- **Goroutines**: Para processamento assíncrono
- **Mutex**: Para controle de concorrência

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT.
