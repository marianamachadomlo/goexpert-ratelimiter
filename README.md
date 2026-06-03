# Rate Limiter (Go)

Middleware de rate limiting para serviços HTTP com persistência em Redis, limitação por IP ou token (`API_KEY: <TOKEN>`) e bloqueio temporário após excesso de requisições.

## Requisitos

- Docker
- Docker Compose

## Execução rápida

1. Copie as variáveis de ambiente:

```bash
cp .env.example .env
```

2. Suba a aplicação e o Redis:

```bash
docker compose up --build
```

3. Teste o serviço:

```bash
curl -i http://localhost:8080/
curl -i -H "API_KEY: <TOKEN>" http://localhost:8080/
```

## Configuração

Todas as configurações são feitas por variáveis de ambiente (arquivo `.env` na raiz):

| Variável | Descrição | Padrão |
|---|---|---|
| `SERVER_PORT` | Porta HTTP da aplicação | `8080` |
| `RATE_LIMIT_IP` | Máximo de requisições/s por IP | `10` |
| `RATE_LIMIT_TOKEN_DEFAULT` | Limite padrão por token | `100` |
| `TOKEN_LIMITS` | JSON com limites por token | `{}` |
| `RATE_WINDOW` | Janela de contagem (ex: `1s`) | `1s` |
| `BLOCK_DURATION` | Tempo de bloqueio após excesso | `5m` |
| `REDIS_ADDR` | Endereço do Redis | `redis:6379` |
| `REDIS_PASSWORD` | Senha do Redis | vazio |
| `REDIS_DB` | Banco lógico do Redis | `0` |

### Exemplo de limites por token

```env
RATE_LIMIT_IP=10
RATE_LIMIT_TOKEN_DEFAULT=20
TOKEN_LIMITS={"premium-token":100,"basic-token":20}
```

Com essa configuração:

- Requisições sem `API_KEY` usam o limite de IP (`10 req/s`)
- Requisições com `API_KEY: premium-token` usam `100 req/s`
- Requisições com `API_KEY: basic-token` usam `20 req/s`
- Tokens não listados usam `RATE_LIMIT_TOKEN_DEFAULT`

**Regra de precedência:** quando o header `API_KEY` está presente, o limitador usa apenas a estratégia de token e ignora o limite de IP.

### Strategy Pattern

A interface `StorageStrategy` define o contrato de persistência:

```go
type StorageStrategy interface {
    IsBlocked(ctx context.Context, key string) (bool, error)
    SetBlock(ctx context.Context, key string, duration time.Duration) error
    IncrementAndCheck(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, count int64, err error)
}
```

Para trocar o Redis por outro mecanismo:

1. Crie uma nova implementação em `internal/limiter/strategy/<nome>/`
2. Injete a nova estratégia em `cmd/server/main.go` no lugar de `redis.New(...)`

A lógica de negócio em `internal/limiter` e o middleware permanecem inalterados.

## Desenvolvimento local (opcional)

Se preferir rodar fora do Docker:

```bash
go test ./... -v
go run ./cmd/server
```

Certifique-se de que o Redis esteja acessível em `REDIS_ADDR`.
