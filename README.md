# Sistema de Leilões com Fechamento Automático

Este projeto implementa um sistema de leilões em Go com funcionalidade de fechamento automático baseado em tempo configurável.

## Funcionalidades

- ✅ Criação de leilões
- ✅ Sistema de lances (bids)
- ✅ Fechamento automático de leilões baseado em tempo
- ✅ Validação de leilões fechados para novos lances
- ✅ API REST para gerenciamento

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

1. **Goroutine de Monitoramento**: Uma goroutine que executa a cada 10 segundos verificando leilões expirados
2. **Controle de Tempo**: Utiliza variáveis de ambiente para configurar o tempo de duração dos leilões
3. **Thread Safety**: Implementa mutex para controle de concorrência
4. **Logging**: Registra automaticamente quando leilões são fechados

### Variáveis de Ambiente

- `AUCTION_INTERVAL`: Duração do leilão (ex: "2m", "30s", "1h")
- `AUCTION_CHECK_INTERVAL`: Intervalo de verificação de leilões expirados (ex: "10s", "30s", "1m")
- `AUCTION_CONTEXT_TIMEOUT`: Timeout para operações de contexto (ex: "30s", "1m", "2m")
- `MONGODB_URI`: URI de conexão com MongoDB
- `MONGODB_DATABASE`: Nome do banco de dados

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
export AUCTION_CHECK_INTERVAL=10s
export AUCTION_CONTEXT_TIMEOUT=30s
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

### Personalizar Intervalo de Verificação

Para alterar a frequência de verificação de leilões expirados, modifique o valor no código:

```go
// Em startAutoCloseRoutine()
ticker := time.NewTicker(time.Second * 10) // Altere este valor
```

### Configurações de Ambiente

Crie um arquivo `.env` na pasta `cmd/auction/`:

```env
MONGODB_URI=mongodb://mongodb:27017
MONGODB_DATABASE=auction_db
AUCTION_INTERVAL=2m
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

- A goroutine verifica leilões a cada 10 segundos
- Para sistemas com muitos leilões, considere ajustar este intervalo
- O sistema usa mutex para thread safety

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
