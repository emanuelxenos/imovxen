# TASK-002: Schema Inicial do Banco e Configuração do sqlc

## Objetivo
Implementar a migration inicial do PostgreSQL com todas as tabelas, índices e triggers da Fase 1, e configurar a geração de código tipado com `sqlc`.

## Dependências
- `TASK-001-infra-docker.md`
- `documento.md` (Seção 5)

## Entregáveis
1. `backend/migrations/0001_init.up.sql`:
   - Extensões: `pgcrypto`, `citext`, `pg_trgm`.
   - Tabelas: `users`, `refresh_tokens`, `properties`, `property_images`, `leads`, `favorites`.
   - Triggers: `fn_sync_user_listings_count` e `trg_sync_user_listings` para controle atômico da cota.
   - Todos os índices otimizados definidos no documento.
2. `backend/migrations/0001_init.down.sql`:
   - Remoção reversível de triggers, tabelas e extensões.
3. `backend/sqlc.yaml`:
   - Configurado para `engine: postgresql`, schema apontando para as migrations, queries em `sql/queries/`, gerando código com `emit_json_tags: true`.
4. Arquivos de query iniciais:
   - `backend/sql/queries/users.sql`
   - `backend/sql/queries/properties.sql`
   - `backend/sql/queries/leads.sql`

## Critério de Aceite
- Migrations rodam no banco de dados sem erro de sintaxe.
- Rodar `sqlc generate` dentro de `backend/` cria os modelos e métodos Go em `backend/internal/db/`.
