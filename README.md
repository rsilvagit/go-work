# go-work

Agregador de vagas de emprego que realiza scraping simultâneo em múltiplas plataformas (LinkedIn, Indeed, Gupy), com filtragem avançada, proteções anti-ban, cache Redis e notificações via Telegram.

## Funcionalidades

- **Multi-plataforma** — Busca vagas no LinkedIn, Indeed e Gupy em paralelo
- **Proteção anti-ban** — User-Agent rotation, headers realistas, rate limiting com jitter, retry com exponential backoff e suporte a proxy
- **Cache Redis** — Evita requests repetidos com TTL configurável (padrão: 1h)
- **Filtragem avançada** — Por tipo de vaga, modelo de trabalho, nível de experiência e região
- **Deduplicação** — Remove vagas duplicadas entre plataformas automaticamente
- **Notificação Telegram** — Envio dos resultados diretamente para um chat/grupo no Telegram
- **Saída formatada** — Exibição em tabela no terminal
- **Extensível** — Adicione novos scrapers implementando a interface `Scraper`

## Tech Stack

- **Go 1.25** — Concorrência nativa com goroutines
- **Redis** — Cache de resultados para reduzir requests
- **goquery** — Parsing de HTML e seletores CSS
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
./go-work -q "desenvolvedor golang" -l "São Paulo"

# Com filtros
./go-work -q "developer" -l "Brasil" -tipo full-time -modelo remoto -nivel senior

# Com cache Redis
./go-work -q "developer" -l "Brasil" -redis-url "redis://localhost:6379"

# Com proxy
./go-work -q "developer" -l "Brasil" -proxy "http://proxy:8080"

# Com notificação Telegram
./go-work -q "developer" -l "Brasil" \
  -telegram-token "$TELEGRAM_TOKEN" \
  -telegram-chat-id "$TELEGRAM_CHAT_ID"
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

## Docker Compose

```bash
# Subir Redis + app
docker-compose up

# Só o Redis (para dev local)
docker-compose up -d redis
./go-work -q "golang" -redis-url "redis://localhost:6379"

# Executar busca via compose
docker-compose run --rm go-work -q "golang developer" -l "São Paulo"
```

## Variáveis de Ambiente

Copie o `.env.example` e preencha com seus dados:

```bash
cp .env.example .env
```

```env
TELEGRAM_TOKEN=seu_token_aqui
TELEGRAM_CHAT_ID=seu_chat_id_aqui
REDIS_URL=redis://localhost:6379
PROXY_URL=
```

## Estrutura do Projeto

```
go-work/
├── cmd/go-work/           # Entrypoint da aplicação
├── internal/
│   ├── httpclient/        # HTTP client com proteções anti-ban
│   ├── cache/             # Cache Redis
│   ├── model/             # Modelo de dados (Job)
│   ├── scraper/           # Scrapers (LinkedIn, Indeed, Gupy)
│   ├── filter/            # Filtros de vagas
│   └── output/            # Writers (Console, Telegram)
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── go.mod
```

## Arquitetura

```
                          ┌──────────────┐
                     ┌───►│   LinkedIn   │
                     │    └──────────────┘
┌─────────┐    ┌─────┴─────┐              ┌───────────┐    ┌─────────┐    ┌─────────┐
│ CLI Flags├───►│ Cache Redis├─miss────────►│ httpclient├───►│ Filtros ├───►│ Output  │
└─────────┘    └─────┬─────┘              └───────────┘    └─────────┘    └─────────┘
                     │    ┌──────────────┐  │ UA Rotation       │          ├─ Console
                     ├───►│   Indeed     │  │ Rate Limit        │          └─ Telegram
                     │    └──────────────┘  │ Retry/Backoff
                     │    ┌──────────────┐  │ Proxy
                     └───►│   Gupy       │  │ Headers
                          └──────────────┘
```

Cada scraper roda em sua própria goroutine. Se um scraper falhar, os demais continuam normalmente. O cache Redis é verificado antes de cada busca — em caso de hit, o scraper é ignorado.

## Proteções Anti-Ban

| Estratégia | Descrição |
|---|---|
| **User-Agent Rotation** | 13 UAs reais (Chrome, Firefox, Edge, Safari) rotacionados por request |
| **Headers Realistas** | Accept, Accept-Language, Sec-Fetch-*, DNT — simula browser real |
| **Rate Limiting** | Delay aleatório (2-5s) entre requests ao mesmo domínio |
| **Retry + Backoff** | Em caso de 429/503, aguarda 2s → 4s → 8s (max 3 tentativas) |
| **Proxy** | Suporte a proxy HTTP/HTTPS para rotação de IP |
| **Cache Redis** | Reduz volume de requests com TTL configurável |

## Contribuindo

1. Fork o projeto
2. Crie sua branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'feat: adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE) para mais informações.
