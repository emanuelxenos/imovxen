# TASK-008: Módulo de Contratos, Carnê de Locação e Baixa Manual Escalonada

## Contexto
Para atender o mercado de aluguéis (incluindo cidades do interior com imóveis populares a partir de R$ 250,00 e imóveis de padrão mais alto), o Imovxen disponibiliza a funcionalidade de **Carnê Digital de Locação**.
O proprietário formaliza o contrato vinculado a um imóvel e um inquilino, gerando automaticamente a grade de parcelas mensais (ex: 12 ou 30 meses). Quando o inquilino paga por fora (dinheiro, Pix pessoal do dono, etc.), o proprietário pode **Dar Baixa Manual** no sistema mediante o pagamento de uma pequena taxa de serviço escalonada (R$ 3,00 / R$ 5,00 / R$ 7,00), que cobre o custo do sistema e gera na hora o **Recibo Oficial de Quitação em PDF**. O proprietário também pode **Reemitir Cobrança** com nova data de vencimento caso o inquilino atrase e prefira pagar pelo sistema.

## Regra de Taxa de Baixa Manual (Escalonamento)
- **Aluguel até R$ 500,00**: Taxa de **R$ 3,00**
- **Aluguel de R$ 500,01 até R$ 1.500,00**: Taxa de **R$ 5,00**
- **Aluguel acima de R$ 1.500,00**: Taxa de **R$ 7,00**

## Dependências
- `documento.md` (Seção 5.3 e Seção 7 - Fase 2)
- PostgreSQL local e pool pgx.

## Entregáveis
1. **Migration PostgreSQL (`backend/migrations/0003_contracts_installments.up.sql`)**:
   - Tabela `contracts`:
     - `id`, `property_id`, `landlord_id`, `tenant_name`, `tenant_email`, `tenant_phone`, `tenant_cpf`, `rent_amount`, `due_day`, `duration_months`, `starts_at`, `status` (`ACTIVE`, `FINISHED`, `CANCELLED`), `created_at`.
   - Tabela `contract_installments`:
     - `id`, `contract_id`, `installment_number`, `due_date`, `amount`, `status` (`PENDING`, `PAID_SYSTEM`, `PAID_MANUAL`, `OVERDUE`), `manual_fee_amount`, `paid_at`, `payment_method`, `notes`, `receipt_url`, `created_at`.
2. **Backend API (`internal/contracts`)**:
   - `POST /api/contracts`: cria o contrato e gera atomicamente todas as parcelas (`contract_installments`) da duração estipulada.
   - `GET /api/me/contracts`: lista os contratos de locação do proprietário logado.
   - `GET /api/contracts/:id`: detalhes do contrato com a lista completa de parcelas do carnê.
   - `POST /api/contracts/installments/:id/manual-fee-checkout`: calcula a taxa (R$ 3, 5 ou 7) com base no valor do aluguel e gera o Pix da taxa no Asaas.
   - `POST /api/contracts/installments/:id/confirm-manual-payment`: confirma a baixa manual da parcela e gera o recibo com dados de quitação.
   - `POST /api/contracts/installments/:id/reissue`: gera nova cobrança com nova data de vencimento para o inquilino.
3. **Frontend SSR & HTMX**:
   - Nova aba na sidebar do Dashboard: **"📜 Contratos & Carnês"** (`/painel/contratos`).
   - Modal para cadastrar novo contrato de locação (seleciona o imóvel, dados do inquilino, valor e dia do vencimento).
   - Tela do **Carnê Mensal**: tabela de parcelas com status colorido (Pendente, Pago por Fora, Em Atraso), botões de **"Dar Baixa Manual"**, **"Ver Recibo"** e **"Reemitir Pix"**.

## Critérios de Aceite
- Ao criar um contrato de 12 meses para um imóvel, o sistema gera as 12 parcelas correspondentes com os vencimentos corretos.
- Aluguel de R$ 250,00 cobra exatamente R$ 3,00 na taxa de baixa manual.
- Aluguel de R$ 800,00 cobra exatamente R$ 5,00 na taxa de baixa manual.
- Aluguel de R$ 2.000,00 cobra exatamente R$ 7,00 na taxa de baixa manual.
- A baixa manual altera o status da parcela para `PAID_MANUAL` e disponibiliza o recibo oficial.
