# IMOVXEN — DIRETRIZES DE PROJETO E IA (AGENTS.md)

Este repositório contém o código do **Imovxen**, um SaaS de Marketplace Imobiliário construído em Go (Backend API e Frontend SSR com Templ + HTMX), PostgreSQL 16 e MinIO/R2.

Qualquer agente de IA (Antigravity, Claude Code, Cursor, Copilot, etc.) ou desenvolvedor humano deve seguir rigorosamente as regras e padrões documentados aqui e em `documento.md`.

---

## 1. Princípios Arquiteturais Fundamentais

1. **Dois Projetos Go Independentes**:
   - `backend/`: API HTTP/JSON pura (Gin + pgx/v5 + sqlc). Não renderiza HTML.
   - `frontend/`: Web Server SSR (Gin + templ + HTMX). Não acessa banco de dados diretamente; consome exclusivamente o backend via HTTP/JSON.
2. **Sem ORM Pesado**:
   - Todo acesso a dados no backend é feito via **PostgreSQL nativo + sqlc**.
   - Toda alteração de schema deve ser feita via migration com arquivos `.up.sql` e `.down.sql` em `backend/migrations/`.
   - O código Go em `backend/internal/db/` é gerado pelo sqlc e **NUNCA** deve ser editado manualmente.
3. **Integridade de Negócio no Banco**:
   - Contagem de anúncios ativos (`listings_used`) é mantida por trigger atômica (`fn_sync_user_listings_count`). Nunca calcule ou incremente isso manualmente sem transação.
4. **Armazenamento de Mídia Eficiente**:
   - Uploads de imagens nunca passam o arquivo binário pelo backend. O backend gera **Presigned URLs** (S3/R2) e o cliente envia direto ao storage.
5. **Autenticação**:
   - JWT stateless curto (15 min) + Refresh Token rotacionado no banco de dados. O frontend cuida da renovação silenciosa em caso de 401.

---

## 2. Estrutura de Pastas e Responsabilidades

```text
imovxen/
├── .agents/
│   ├── skills/              # Guias procedurais especializados (sqlc, templ, migrations)
│   └── taskcards/           # Tarefas modulares e incrementais de desenvolvimento
├── backend/
│   ├── cmd/api/main.go      # Ponto de entrada da API
│   ├── internal/
│   │   ├── auth/            # JWT, senhas, tokens
│   │   ├── users/           # Domínio de usuários e planos
│   │   ├── properties/      # Domínio de imóveis e busca
│   │   ├── storage/         # Integração S3 / Presigned URLs
│   │   └── db/              # Pool pgx e arquivos gerados por sqlc (não editar)
│   ├── migrations/          # Scripts SQL golang-migrate
│   ├── sql/queries/         # Queries SQL de entrada para o sqlc
│   └── sqlc.yaml
├── frontend/
│   ├── cmd/web/main.go      # Ponto de entrada do Frontend SSR
│   ├── internal/
│   │   ├── api/             # Cliente HTTP tipado com auto-refresh de JWT
│   │   ├── handlers/        # Rotas HTTP e renderização de views
│   │   └── views/           # Componentes templ (.templ)
│   └── static/              # CSS, HTMX, imagens estáticas
├── docker-compose.yml       # Postgres 16 + MinIO local
└── documento.md             # Especificação e contrato técnico do produto
```

---

## 3. Padrões de Código e Convenções

### Backend (Go)
- **Estruturação por Camadas**: `handler` (Gin) -> `service` (negócio) -> `repository` (sqlc/db).
- **Tratamento de Erros**: Sempre use erros explícitos e sentinels em `internal/errs`. Nunca ignore erros com `_`.
- **Logs**: Use a biblioteca nativa `log/slog` com campos estruturados em JSON e propagação de `request_id`.
- **Validação de Entrada**: Use tags `binding:"required"` do Gin / `go-playground/validator` nos structs DTO.

### Frontend (Go + Templ + HTMX)
- **Componentização com `templ`**: Todo HTML deve ser escrito em arquivos `.templ` tipados. Gere o código com `templ generate`.
- **Interatividade**: Prefira atributos HTMX (`hx-get`, `hx-post`, `hx-target`, `hx-swap`) a scripts JavaScript manuais.

---

## 4. Como Trabalhar com Tarefas (Taskcards)

- Todas as tarefas a executar estão descritas em `.agents/taskcards/`.
- Cada Taskcard possui contexto, dependências, passos claros e critério de aceite.
- Antes de iniciar uma tarefa, leia o Taskcard correspondente e a especificação em `documento.md`.
