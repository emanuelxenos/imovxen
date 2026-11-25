---
name: imovxen-backend-guide
description: >-
  Guia para desenvolvimento no backend Go do Imovxen. Use sempre que for criar
  ou alterar rotas Gin, queries sqlc, migrations do PostgreSQL, regras de negócio ou serviços.
---

# Backend Development Guide — Imovxen

Este guia descreve os padrões para implementação no projeto `backend/`.

## 1. Fluxo de Criação de Entidades e Banco de Dados

1. **Criar a Migration**:
   - Adicione os arquivos `backend/migrations/NNNN_<nome>.up.sql` e `.down.sql`.
   - Sempre utilize tipos explícitos (`TIMESTAMPTZ`, `UUID`, `NUMERIC(12,2)`, `CITEXT`).
2. **Definir Queries para o `sqlc`**:
   - Adicione os comandos SQL em `backend/sql/queries/<dominio>.sql`.
   - Anote cada query com o nome e tipo:
     ```sql
     -- name: GetPropertyByID :one
     SELECT * FROM properties WHERE id = $1 LIMIT 1;

     -- name: ListActiveProperties :many
     SELECT * FROM properties 
     WHERE status = 'ACTIVE' 
     ORDER BY created_at DESC 
     LIMIT $1 OFFSET $2;
     ```
3. **Gerar o Código Go**:
   - Execute o gerador do sqlc (`sqlc generate`). Nunca altere o código em `internal/db/` na mão.

## 2. Padrão de Camadas

```text
HTTP Request -> Handler (Gin) -> Service (Regra/Validação) -> Repository (sqlc.Queries)
```

- **Handler (`internal/<dominio>/handler.go`)**:
  - Faz bind e valida o DTO (`c.ShouldBindJSON`).
  - Extrai identidade do usuário via contexto (`c.MustGet("user_id")`).
  - Chama o service correspondente e formata a resposta JSON.
- **Service (`internal/<dominio>/service.go`)**:
  - Executa as regras de validação de negócio (ex: verificar cota em `users.listings_used`).
  - Gerencia transações (`pgx.Tx`) quando mais de uma query precisa ser atômica.
- **DTOs (`internal/<dominio>/dto.go`)**:
  - Structs com tags `json` e tags de validação `binding`.
