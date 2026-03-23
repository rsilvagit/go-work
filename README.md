# go-work

Agregador de vagas de emprego que consome a API pública da Gupy, com filtragem avançada, proteções anti-ban e notificações via Discord. Execução automatizada via GitHub Actions (cron diário gratuito).

> **[Documentação Técnica Completa](https://rsilvagit.github.io/go-work/)** — Arquitetura, decisões técnicas, padrões de código, pipeline de dados e deploy.

## Funcionalidades

- **API Gupy** — Consome a API JSON pública (`employability-portal.gupy.io`) — sem parsing de HTML, dados estruturados
- **Multi-query** — Busca múltiplas stacks em paralelo (`golang,python,c#`)
- **Proteção anti-ban** — User-Agent rotation, headers realistas, rate limiting com jitter, retry com exponential backoff e suporte a proxy
- **Filtragem avançada** — Por tipo de vaga, modelo de trabalho (`remoto,hibrido`), nível e região. Suporta múltiplos valores por vírgula
- **Apenas vagas novas** — Filtra automaticamente vagas postadas nas últimas 24h (configurável), eliminando duplicatas entre execuções
- **Deduplicação** — Remove vagas duplicadas dentro da mesma execução
- **Notificação Discord** — Envio via Webhook com formatação Markdown e chunking automático (limite 2000 chars)
- **Notificação Telegram** — Envio dos resultados diretamente para um chat/grupo
- **Saída formatada** — Exibição em tabela no terminal
- **Cron GitHub Actions** — Execução automática diária às 12h UTC / 9h BRT (gratuito)
- **Auto-run on push** — Executa automaticamente a cada push em `main`
- **Extensível** — Adicione novos scrapers implementando a interface `Scraper`

## Tech Stack

- **Go 1.25** — Concorrência nativa com goroutines
- **GitHub Actions** — Cron diário + CI/CD
- **Docker + Compose** — Para desenvolvimento local

## Pré-requisitos

- Go 1.25+ ou Docker

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

# Múltiplas stacks
./go-work -q "golang,python,c#"

# Com filtros (suporta múltiplos valores por vírgula)
./go-work -q "golang,python" -modelo "remoto,hibrido" -nivel senior

# Com notificação Discord
./go-work -q "golang,python,c#" -modelo "remoto,hibrido" \
  -discord-webhook "$DISCORD_WEBHOOK_URL"

# Com notificação Telegram
./go-work -q "developer" \
  -telegram-token "$TELEGRAM_TOKEN" \
  -telegram-chat-id "$TELEGRAM_CHAT_ID"

# Com proxy
./go-work -q "developer" -proxy "http://proxy:8080"
```

### Flags

| Flag | Descrição | Padrão |
|------|-----------|--------|
| `-q` | Termos de busca separados por vírgula (obrigatório) | — |
| `-timeout` | Timeout por scraper | `30s` |
| `-tipo` | Tipo de vaga (`full-time`, `part-time`, `estagio`, `freelance`) | — |
| `-modelo` | Modelo de trabalho (`remoto`, `hibrido`, `presencial`) | — |
| `-nivel` | Nível (`junior`, `pleno`, `senior`) | — |
| `-regiao` | Filtro por região/cidade | — |
| `-l` | Localização para filtrar na API (ex: `São Paulo`) | — |
| `-redis-url` | URL do Redis para cache (ex: `redis://localhost:6379`) | — |
| `-cache-ttl` | TTL do cache de resultados | `1h` |
| `-proxy` | URL do proxy HTTP/HTTPS | — |
| `-min-delay` | Delay mínimo entre requests (anti-ban) | `2s` |
| `-max-delay` | Delay máximo entre requests (anti-ban) | `5s` |
| `-telegram-token` | Token do Bot Telegram | — |
| `-telegram-chat-id` | Chat ID do Telegram | — |
| `-discord-webhook` | URL do Webhook Discord | — |

Flags e filtros suportam múltiplos valores separados por vírgula (ex: `-q "golang,python"`, `-modelo "remoto,hibrido"`).

## Variáveis de Ambiente

Copie o `.env.example` e preencha com seus dados:

```bash
cp .env.example .env
```

```env
# Busca
SEARCH_QUERY=golang,python,c#
SEARCH_LOCATION=São Paulo
SEARCH_MODELO=remoto,hibrido
SEARCH_TIPO=full-time
SEARCH_NIVEL=senior
SEARCH_REGIAO=São Paulo

# Notificações
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/xxx/yyy
TELEGRAM_TOKEN=seu_token_aqui
TELEGRAM_CHAT_ID=seu_chat_id_aqui

# Cache e Proxy (opcional)
REDIS_URL=redis://localhost:6379
PROXY_URL=http://proxy:8080
```

Todas as flags da CLI possuem fallback para variáveis de ambiente, permitindo execução 100% via env vars (ideal para GitHub Actions cron).

## Estrutura do Projeto

```
go-work/
├── cmd/go-work/           # Entrypoint da aplicação
├── internal/
│   ├── cache/             # Cache Redis (opcional)
│   ├── httpclient/        # HTTP client com proteções anti-ban
│   ├── model/             # Modelo de dados (Job)
│   ├── scraper/           # Scraper Gupy (API JSON)
│   ├── filter/            # Filtros de vagas (inclui filtro de 24h)
│   └── output/            # Writers (Console, Telegram, Discord)
├── .github/workflows/     # Cron + CI (GitHub Actions)
├── docker-compose.yml     # Redis para desenvolvimento local
├── Dockerfile
├── .env.example
└── go.mod
```

## Arquitetura

```
┌─────────┐    ┌───────────┐    ┌────────────────┐    ┌──────────┐    ┌─────────┐    ┌─────────┐
│ CLI/Env  ├───►│ httpclient├───►│ Gupy API (JSON)├───►│ Dedup    ├───►│ Filtros ├───►│ Output  │
└─────────┘    └─────┬─────┘    └───────┬────────┘    └──────────┘    └────┬────┘    └─────────┘
                     │                  │                                  │           ├─ Console
                     │ UA Rotation      │ Cache Redis                      │ MaxAge    ├─ Discord
                     │ Rate Limit       │ (opcional)                       │ Tipo      └─ Telegram
                     │ Retry/Backoff                                       │ Modelo
                     │ Proxy                                               │ Nível
                                                                           │ Região
```

Cada termo de busca gera uma goroutine separada. Os resultados são combinados, deduplicados por URL (ou título+empresa), filtrados por idade (últimas 24h) e critérios do usuário, e então enviados para os canais configurados.

## Proteções Anti-Ban

| Estratégia | Descrição |
|---|---|
| **User-Agent Rotation** | 13 UAs reais (Chrome, Firefox, Edge, Safari) rotacionados por request |
| **Headers Realistas** | Accept, Accept-Language, Sec-Fetch-*, DNT — simula browser real |
| **Rate Limiting** | Delay aleatório (2-5s) entre requests ao mesmo domínio |
| **Retry + Backoff** | Em caso de 429/503, aguarda 2s → 4s → 8s (max 3 tentativas) |
| **Proxy** | Suporte a proxy HTTP/HTTPS para rotação de IP |
| **Cache Redis** | Reduz volume de requests com TTL configurável |

## Filtro de Vagas Recentes

Por padrão, apenas vagas publicadas nas **últimas 24 horas** são retornadas. Isso evita notificações duplicadas entre execuções diárias — cada dia traz somente vagas novas.

O filtro usa o campo `publishedDate` da API da Gupy. Vagas sem data de publicação passam normalmente.

> Para ajustar o período, altere o valor de `MaxAge` no `filter.Options` ao chamar `filter.Apply()`.

## Cache Redis (Opcional)

O cache Redis reduz chamadas repetidas à API durante a mesma janela de tempo:

```bash
# Com Redis local
./go-work -q "golang" -redis-url "redis://localhost:6379"

# TTL customizado
./go-work -q "golang" -redis-url "redis://localhost:6379" -cache-ttl 2h
```

- **Chave:** `gowork:{scraper}:{sha256(scraper:query:location)}`
- **TTL padrão:** 1 hora
- **Fallback:** se o Redis estiver indisponível, a aplicação continua normalmente sem cache

Para desenvolvimento local com Docker:

```bash
docker-compose up -d   # sobe o Redis na porta 6379
```

## GitHub Actions (Cron + CI)

O workflow em `.github/workflows/deploy.yml` faz tudo:

- **Cron diário** — executa `go-work` automaticamente às 12h UTC (9h BRT)
- **Push em `main`** — build + execução a cada push
- **Manual** — pode ser disparado manualmente via `workflow_dispatch`

### Setup

1. No GitHub, vá em **Settings → Secrets and variables → Actions**
2. Adicione os **Repository Secrets** abaixo
3. Pronto — o cron roda diariamente e envia as vagas para Discord/Telegram

### GitHub Secrets

| Secret | Descrição |
|---|---|
| `SEARCH_QUERY` | Termos de busca separados por vírgula (obrigatório). Ex: `golang,python,c#` |
| `DISCORD_WEBHOOK_URL` | URL do Webhook Discord |
| `SEARCH_MODELO` | Modelo de trabalho (opcional). Ex: `remoto,hibrido` |
| `SEARCH_TIPO` | Tipo de vaga (opcional). Ex: `full-time` |
| `SEARCH_NIVEL` | Nível (opcional). Ex: `senior` |
| `SEARCH_REGIAO` | Região (opcional) |
| `TELEGRAM_TOKEN` | Token do Bot Telegram (opcional) |
| `TELEGRAM_CHAT_ID` | Chat ID do Telegram (opcional) |
| `SEARCH_LOCATION` | Localização para filtrar na API (opcional). Ex: `São Paulo` |
| `PROXY_URL` | URL do proxy (opcional) |
| `REDIS_URL` | URL do Redis para cache (opcional). Ex: `redis://localhost:6379` |

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
