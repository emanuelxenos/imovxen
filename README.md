# Guia de Engenharia e Manutenção com IA — Imovxen

Bem-vindo ao repositório do **Imovxen**! Este projeto foi estruturado seguindo os métodos mais modernos de engenharia de software auxiliada por Inteligência Artificial (AI-Assisted Development / Agentic Workflows).

---

## 1. Visão Geral da Arquitetura do Repositório

O projeto adota uma arquitetura em **Monorepo com dois projetos Go desacoplados**:
1. **`backend/`**: API REST/JSON construída em Go com Gin, PostgreSQL 16, `pgx/v5`, queries tipadas via `sqlc` e armazenamento S3/R2 com Presigned URLs.
2. **`frontend/`**: Servidor Web SSR construído em Go com Gin, templates tipados em tempo de compilação via `a-h/templ` e reatividade de interface via HTMX. O frontend consome exclusivamente o backend via HTTP.

---

## 2. Como Funciona a Estrutura de IA (`.agents/` e `AGENTS.md`)

Qualquer agente de IA (Antigravity, Cursor, Copilot, Claude Code) ou colega de equipe humano encontra no projeto um ecossistema claro de diretrizes:

### A. [AGENTS.md](file:///c:/Users/Emanuel/Desktop/imovxen/AGENTS.md)
* É o **ponto de entrada universal** lido automaticamente pelos agentes de IA.
* Define os princípios inegociáveis: sem ORM pesado, integridade de cota garantida por triggers no banco, uploads direto para o storage e divisão de responsabilidades.

### B. Skills de Especialização (`.agents/skills/`)
As Skills ensinam os padrões exatos de código que a IA deve seguir para não inventar padrões estranhos:
* [backend-guide](file:///c:/Users/Emanuel/Desktop/imovxen/.agents/skills/backend-guide/SKILL.md): Padrão de camadas (`handler -> service -> repository/sqlc`), fluxo de migrations e queries `.sql`.
* [frontend-guide](file:///c:/Users/Emanuel/Desktop/imovxen/.agents/skills/frontend-guide/SKILL.md): Padrão de views com `templ`, estrutura de componentes, partials HTMX e cliente HTTP com auto-refresh de JWT.

### C. Taskcards Modulares (`.agents/taskcards/`)
As tarefas foram divididas em módulos atômicos, dependentes e testáveis:
* [TASK-001: Infraestrutura Básica e Docker](file:///c:/Users/Emanuel/Desktop/imovxen/.agents/taskcards/TASK-001-infra-docker.md)
* [TASK-002: Schema do Banco de Dados e sqlc](file:///c:/Users/Emanuel/Desktop/imovxen/.agents/taskcards/TASK-002-database-schema-sqlc.md)
* [TASK-003: Backend Base e Autenticação JWT](file:///c:/Users/Emanuel/Desktop/imovxen/.agents/taskcards/TASK-003-backend-auth.md)
* [TASK-004: CRUD de Imóveis, Filtros e Leads](file:///c:/Users/Emanuel/Desktop/imovxen/.agents/taskcards/TASK-004-backend-properties-leads.md)
* [TASK-005: Upload com Presigned URLs](file:///c:/Users/Emanuel/Desktop/imovxen/.agents/taskcards/TASK-005-backend-storage-presigned.md)
* [TASK-006: Frontend SSR com Templ e HTMX](file:///c:/Users/Emanuel/Desktop/imovxen/.agents/taskcards/TASK-006-frontend-ssr-htmx.md)

---

## 3. Como Trabalhar em Equipe ou Delegar Tarefas para IAs

1. **Passar para outra pessoa/IA**:
   * Basta pedir para a IA ler o arquivo [AGENTS.md](file:///c:/Users/Emanuel/Desktop/imovxen/AGENTS.md) e o Taskcard que deseja executar (por exemplo: *"Execute os passos descritos em .agents/taskcards/TASK-001-infra-docker.md"*).
   * A IA já saberá exatamente o que produzir, onde colocar os arquivos e como validar os critérios de aceite.
2. **Manutenção Futura**:
   * Se precisar adicionar uma nova funcionalidade (ex: Favoritos ou Fase 2 de Contratos), basta criar um novo Taskcard (`TASK-007-contratos.md`), atualizar o `documento.md` e instruir o agente a implementar.
