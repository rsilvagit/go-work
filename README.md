# go-work

Agregador de vagas de emprego que consome a API pública da Gupy, com filtragem avançada, proteções anti-ban, cache Redis e notificações via Telegram e Discord. Execução automatizada via GitHub Actions (cron diário gratuito).

## Funcionalidades

- **API Gupy** — Consome a API JSON pública (`employability-portal.gupy.io`) — sem parsing de HTML, dados estruturados
- **Proteção anti-ban** — User-Agent rotation, headers realistas, rate limiting com jitter, retry com exponential backoff e suporte a proxy
- **Cache Redis** — Evita requests repetidos com TTL configurável (padrão: 1h)
- **Filtragem avançada** — Por tipo de vaga, modelo de trabalho, nível de experiência e região
- **Deduplicação** — Remove vagas duplicadas automaticamente
- **Notificação Telegram** — Envio dos resultados diretamente para um chat/grupo no Telegram
- **Notificação Discord** — Envio via Webhook com formatação Markdown e chunking automático (limite 2000 chars)
- **Saída formatada** — Exibição em tabela no terminal
- **Cron GitHub Actions** — Execução automática diária às 12h UTC / 9h BRT (gratuito)
- **Auto-run on push** — Executa automaticamente a cada push em `main`
- **Extensível** — Adicione novos scrapers implementando a interface `Scraper`

## Tech Stack

- **Go 1.25** — Concorrência nativa com goroutines
- **Redis** — Cache de resultados para reduzir requests
- **Docker + Compose** — Orquestração de app + Redis

## Pré-requisitos

- Go 1.25+ ou Docker
- Redis (opcional — sem ele a app funciona normalmente, apenas sem cache)

## Instalação

```bash
git clone https://github.com/rsilvagit/go-work.git
cd go-work
go build -o go-work ./cmd/go-work
```

## Uso

```bash
# Busca básica
./go-work -q "golang"

# Com filtros
./go-work -q "developer" -modelo remoto -nivel senior -tipo full-time

# Com cache Redis
./go-work -q "developer" -redis-url "redis://localhost:6379"

# Com proxy
./go-work -q "developer" -proxy "http://proxy:8080"

# Com notificação Telegram
./go-work -q "developer" \
  -telegram-token "$TELEGRAM_TOKEN" \
  -telegram-chat-id "$TELEGRAM_CHAT_ID"

# Com notificação Discord
./go-work -q "developer" \
  -discord-webhook "$DISCORD_WEBHOOK_URL"

# Com Telegram + Discord
./go-work -q "golang developer" -modelo remoto \
  -telegram-token "$TELEGRAM_TOKEN" \
  -telegram-chat-id "$TELEGRAM_CHAT_ID" \
  -discord-webhook "$DISCORD_WEBHOOK_URL"
```

### Flags

| Flag | Descrição | Padrão |
|------|-----------|--------|
| `-q` | Termo de busca (obrigatório) | — |
| `-l` | Localização | — |
| `-timeout` | Timeout por scraper | `30s` |
| `-tipo` | Tipo de vaga (`full-time`, `part-time`, `estagio`, `freelance`) | — |
| `-modelo` | Modelo de trabalho (`remoto`, `hibrido`, `presencial`) | — |
| `-nivel` | Nível (`junior`, `pleno`, `senior`) | — |
| `-regiao` | Filtro por região/cidade | — |
| `-redis-url` | URL do Redis para cache | — |
| `-cache-ttl` | TTL do cache | `1h` |
| `-proxy` | URL do proxy HTTP/HTTPS | — |
| `-min-delay` | Delay mínimo entre requests (anti-ban) | `2s` |
| `-max-delay` | Delay máximo entre requests (anti-ban) | `5s` |
| `-telegram-token` | Token do Bot Telegram | — |
| `-telegram-chat-id` | Chat ID do Telegram | — |
| `-discord-webhook` | URL do Webhook Discord | — |

## Docker Compose

```bash
# Subir Redis + app
docker-compose up

# Só o Redis (para dev local)
docker-compose up -d redis
./go-work -q "golang" -redis-url "redis://localhost:6379"

# Executar busca via compose
docker-compose run --rm go-work -q "golang developer"
```

## Variáveis de Ambiente

Copie o `.env.example` e preencha com seus dados:

```bash
cp .env.example .env
```

```env
# Busca
SEARCH_QUERY=golang developer
SEARCH_LOCATION=Brasil
SEARCH_TIPO=full-time
SEARCH_MODELO=remoto
SEARCH_NIVEL=senior
SEARCH_REGIAO=São Paulo

# Notificações
TELEGRAM_TOKEN=seu_token_aqui
TELEGRAM_CHAT_ID=seu_chat_id_aqui
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/xxx/yyy

# Infraestrutura (opcional)
REDIS_URL=redis://localhost:6379
PROXY_URL=
```

Todas as flags da CLI possuem fallback para variáveis de ambiente, permitindo execução 100% via env vars (ideal para GitHub Actions cron).

## Estrutura do Projeto

```
go-work/
├── cmd/go-work/           # Entrypoint da aplicação
├── internal/
│   ├── httpclient/        # HTTP client com proteções anti-ban
│   ├── cache/             # Cache Redis
│   ├── model/             # Modelo de dados (Job)
│   ├── scraper/           # Scraper Gupy (API JSON)
│   ├── filter/            # Filtros de vagas
│   └── output/            # Writers (Console, Telegram, Discord)
├── .github/workflows/     # CI/CD + Cron (GitHub Actions)
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── go.mod
```

## Arquitetura

```
┌─────────┐    ┌─────────────┐              ┌───────────┐    ┌─────────┐    ┌─────────┐
│ CLI Flags├───►│ Cache Redis ├─miss────────►│ httpclient├───►│ Filtros ├───►│ Output  │
└─────────┘    └──────┬──────┘              └─────┬─────┘    └─────────┘    └─────────┘
                      │                           │                          ├─ Console
                      │    ┌────────────────┐     │ UA Rotation              ├─ Telegram
                      └───►│ Gupy API (JSON)│     │ Rate Limit               └─ Discord

                           └────────────────┘     │ Retry/Backoff
                                                  │ Proxy
                                                  │ Headers
```

O scraper consome a API JSON pública da Gupy, sem necessidade de parsear HTML. O cache Redis é verificado antes de cada busca — em caso de hit, a API não é chamada.

## Proteções Anti-Ban

| Estratégia | Descrição |
|---|---|
| **User-Agent Rotation** | 13 UAs reais (Chrome, Firefox, Edge, Safari) rotacionados por request |
| **Headers Realistas** | Accept, Accept-Language, Sec-Fetch-*, DNT — simula browser real |
| **Rate Limiting** | Delay aleatório (2-5s) entre requests ao mesmo domínio |
| **Retry + Backoff** | Em caso de 429/503, aguarda 2s → 4s → 8s (max 3 tentativas) |
| **Proxy** | Suporte a proxy HTTP/HTTPS para rotação de IP |
| **Cache Redis** | Reduz volume de requests com TTL configurável |

## GitHub Actions (Cron + CI/CD)

O workflow em `.github/workflows/deploy.yml` faz tudo:

- **Cron diário** — executa `go-work` automaticamente às 12h UTC (9h BRT)
- **Push em `main`** — build + execução a cada push
- **Manual** — pode ser disparado manualmente via `workflow_dispatch`

### Setup

1. No GitHub, vá em **Settings → Secrets and variables → Actions**
2. Adicione os **Repository Secrets** abaixo
3. Pronto — o cron roda diariamente e envia as vagas para Discord/Telegram

### GitHub Secrets (obrigatório)

| Secret | Descrição |
|---|---|
| `SEARCH_QUERY` | Termo de busca (obrigatório) |
| `TELEGRAM_TOKEN` | Token do Bot Telegram |
| `TELEGRAM_CHAT_ID` | Chat ID do Telegram |
| `DISCORD_WEBHOOK_URL` | URL do Webhook Discord |
| `SEARCH_LOCATION` | Localização (opcional) |
| `SEARCH_TIPO` | Tipo de vaga (opcional) |
| `SEARCH_MODELO` | Modelo de trabalho (opcional) |
| `SEARCH_NIVEL` | Nível (opcional) |
| `SEARCH_REGIAO` | Região (opcional) |
| `PROXY_URL` | URL do proxy (opcional) |

### Execução manual

No GitHub → **Actions** → **Job Search Cron** → **Run workflow** para testar a qualquer momento.

## Contribuindo

1. Fork o projeto
2. Crie sua branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'feat: adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE) para mais informações.
