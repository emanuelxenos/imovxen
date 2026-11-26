# TASK-007: Módulo Financeiro, Pagamentos Asaas e Upgrade de Cota

## Contexto
O sistema possui cota de anúncios (`listing_quota`) e contagem ativa (`listings_used`). Quando um usuário atinge o limite do plano gratuito (2 anúncios), ele é bloqueado ao tentar cadastrar novos imóveis.
Este taskcard implementa o ecossistema de pagamentos integrado ao gateway **Asaas** (modo Sandbox e Produção), permitindo ao anunciante adquirir pacotes de anúncios adicionais ou assinar planos via Pix (com QR Code dinâmico e Copia-e-Cola). Ao receber a confirmação de pagamento (seja via Webhook do Asaas ou simulação em ambiente local), o sistema credita as novas cotas no banco de dados atomicamente e atualiza o extrato financeiro.

## Dependências
- `documento.md` (Seção 5.4 e Seção 7 - Fase 3)
- `backend/migrations/` e PostgreSQL local.

## Entregáveis
1. **Migration PostgreSQL**:
   - Criação da tabela `payments` (`id`, `user_id`, `kind`, `amount`, `status`, `method`, `gateway_id`, `pix_code`, `pix_qr_base64`, `paid_at`, `created_at`).
   - Trigger ou função de incremento atômico de `listing_quota` quando `payments.status` mudar para `PAID` com `kind = 'LISTING_PACK'`.
2. **Backend API (`internal/payments`)**:
   - Cliente Asaas configurável com chave de API e URL base sandbox (`https://sandbox.asaas.com/api/v3`).
   - `POST /api/payments/checkout`: gera cobrança Pix no Asaas para pacote de anúncios (ex: Pacote +5 anúncios por R$ 49,90 ou Pacote +10 anúncios por R$ 89,90).
   - `POST /api/payments/webhook`: endpoint que recebe evento `PAYMENT_RECEIVED` ou `PAYMENT_CONFIRMED` do Asaas e atualiza o pagamento e cota do usuário.
   - `GET /api/me/payments`: extrato financeiro do usuário autenticado.
3. **Frontend SSR & HTMX**:
   - Nova aba na sidebar do Dashboard: **"Financeiro & Planos"** (`/painel/financeiro`).
   - Cards de pacotes de anúncio com botão de checkout Pix instantâneo.
   - Modal com QR Code Pix gerado, código copia-e-cola e polling HTMX leve verificando status do pagamento.
   - Listagem do histórico de transações/recibos.

## Critérios de Aceite
- Usuário consegue visualizar seu plano atual e cotas restantes na aba "Financeiro".
- Ao selecionar um pacote de anúncios, uma cobrança Pix real é gerada no Asaas com QR Code.
- A confirmação de pagamento (via webhook ou simulação de aprovação) incrementa `listing_quota` do usuário imediatamente.
- Usuário pode cadastrar novos imóveis após a liberação da cota sem receber erro 403.
