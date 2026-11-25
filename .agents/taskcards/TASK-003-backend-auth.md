# TASK-003: Estrutura Base do Backend e Autenticação JWT

## Objetivo
Estruturar o projeto `backend/` em Go, configurar o pool de conexão `pgx/v5`, e implementar o fluxo completo de autenticação (registro, login, refresh token, middleware de proteção).

## Dependências
- `TASK-002-database-schema-sqlc.md`
- `documento.md` (Seção 6.1 e 10)

## Entregáveis
1. Inicialização do módulo Go em `backend/go.mod`.
2. Servidor Gin básico em `backend/cmd/api/main.go` com rota de health check (`GET /api/health`).
3. Módulo `backend/internal/auth/`:
   - Hash de senhas seguro (bcrypt / argon2id).
   - Geração e validação de JWT curto (15 min).
   - Geração e rotação de Refresh Token armazenado na tabela `refresh_tokens`.
   - Middleware de autenticação Gin que injeta `userID` no contexto.
4. Handlers e rotas:
   - `POST /api/auth/register`
   - `POST /api/auth/login`
   - `POST /api/auth/refresh`
   - `GET /api/me` (protegido por JWT)

## Critério de Aceite
- Usuário consegue se registrar e fazer login recebendo `access_token` e `refresh_token`.
- Rota protegida `/api/me` devolve os dados do usuário autenticado e status da cota de anúncios.
- Token expirado pode ser renovado com sucesso via `/api/auth/refresh`.
