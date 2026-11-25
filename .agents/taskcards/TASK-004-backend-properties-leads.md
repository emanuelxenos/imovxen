# TASK-004: CRUD de Imóveis, Filtros de Busca e Leads

## Objetivo
Implementar os endpoints de gestão de imóveis (com validação de cota ativa), endpoint público de busca com filtros e captura de leads.

## Dependências
- `TASK-003-backend-auth.md`
- `documento.md` (Seção 6.2 e 6.3)

## Entregáveis
1. Módulo `backend/internal/properties/`:
   - `POST /api/properties`: Valida se `listings_used < listing_quota`; cria imóvel como `ACTIVE`.
   - `GET /api/properties`: Busca pública paginada com suporte a filtros (`city`, `purpose`, `type`, `min_price`, `max_price`, `bedrooms`).
   - `GET /api/properties/:id`: Retorna dados completos do imóvel e imagens.
   - `PATCH /api/properties/:id`: Atualização de imóvel (apenas o dono).
   - `DELETE /api/properties/:id`: Exclusão ou desativação de imóvel (apenas o dono).
2. Módulo `backend/internal/leads/`:
   - `POST /api/properties/:id/leads`: Envio de mensagem pública de contato para o anunciante.
   - `GET /api/me/leads`: Lista de leads recebidos pelos imóveis do usuário autenticado.

## Critério de Aceite
- Usuário não consegue ultrapassar o limite gratuito de anúncios (`listing_quota`).
- A busca pública filtra corretamente e retorna o payload com contagem total e paginação.
- Visitante consegue registrar lead e o anunciante visualiza no `/api/me/leads`.
