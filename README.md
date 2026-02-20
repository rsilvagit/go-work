# go-work

Agregador de vagas de emprego que realiza scraping simultâneo em múltiplas plataformas (LinkedIn, Indeed, Gupy), com filtragem avançada e notificações via Telegram.

## Funcionalidades

- **Multi-plataforma** — Busca vagas no LinkedIn, Indeed e Gupy em paralelo
- **Filtragem avançada** — Por tipo de vaga, modelo de trabalho, nível de experiência e região
- **Deduplicação** — Remove vagas duplicadas entre plataformas automaticamente
- **Notificação Telegram** — Envio dos resultados diretamente para um chat/grupo no Telegram
- **Saída formatada** — Exibição em tabela no terminal
- **Extensível** — Adicione novos scrapers implementando a interface `Scraper`

## Tech Stack

- **Go 1.25** — Concorrência nativa com goroutines
- **goquery** — Parsing de HTML e seletores CSS
- **Docker** — Build multi-stage para imagem mínima

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
./go-work -q "desenvolvedor golang" -l "São Paulo"

# Com filtros
./go-work -q "developer" -l "Brasil" -tipo full-time -modelo remoto -nivel senior

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
| `-telegram-token` | Token do Bot Telegram | — |
| `-telegram-chat-id` | Chat ID do Telegram | — |

## Docker

```bash
docker build -t go-work .
docker run --rm --env-file .env go-work -q "golang developer" -l "São Paulo"
```

## Variáveis de Ambiente

Copie o `.env.example` e preencha com seus dados:

```bash
cp .env.example .env
```

```env
TELEGRAM_TOKEN=seu_token_aqui
TELEGRAM_CHAT_ID=seu_chat_id_aqui
```

## Estrutura do Projeto

```
go-work/
├── cmd/go-work/        # Entrypoint da aplicação
├── internal/
│   ├── model/          # Modelo de dados (Job)
│   ├── scraper/        # Scrapers (LinkedIn, Indeed, Gupy)
│   ├── filter/         # Filtros de vagas
│   └── output/         # Writers (Console, Telegram)
├── Dockerfile
├── .env.example
└── go.mod
```

## Arquitetura

```
[CLI Flags] → [Scrapers em paralelo] → [Deduplicação] → [Filtros] → [Output Writers]
                  ├── LinkedIn
                  ├── Indeed
                  └── Gupy
```

Cada scraper roda em sua própria goroutine com timeout independente. Se um scraper falhar, os demais continuam normalmente.

## Contribuindo

1. Fork o projeto
2. Crie sua branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'feat: adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE) para mais informações.
