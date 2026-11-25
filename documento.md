================================================================================
DOCUMENTO DE DESENVOLVIMENTO — SISTEMA IMOBILIÁRIO SaaS
Versão: 1.1 (Revisada com Melhorias de Arquitetura e Integridade)
Data: 25/11/2026
Stack: Go (backend + frontend) · PostgreSQL 16 · Asaas
Arquitetura: Dois projetos Go independentes (API + Web), HTTP/JSON
================================================================================

--------------------------------------------------------------------------------
1. VISÃO GERAL DO PRODUTO
--------------------------------------------------------------------------------
Marketplace imobiliário público (SaaS) onde qualquer usuário pode anunciar
imóveis para venda ou aluguel. Diferenciais previstos para fases futuras:
contratos digitais com assinatura eletrônica e pagamento de aluguel pelo
sistema.

1.1 Princípios de arquitetura
- Dois projetos Go independentes (backend/ e frontend/) com go.mod próprios.
- Contrato via HTTP/JSON — o front nunca acessa o banco.
- Multi-tenant desde o dia 1 — cada usuário tem plano e cota de anúncios.
- Monetização preparada — modelo freemium (N anúncios grátis, depois cobrança
  por anúncio ou assinatura).
- Performance — Go puro, pgx/v5 + sqlc (queries tipadas), sem ORM pesado, SSR com templ + HTMX.
- Eficiência de I/O — uploads pesados via Direct-to-Storage (Presigned URLs S3/R2),
  evitando streaming duplo de fotos pelo backend.

1.2 Fora de escopo (por enquanto)
- Pagamento de compra de imóvel via checkout (exige cartório/financiamento —
  tratado apenas como sinal/reserva em fase futura).
- App mobile nativo.
- Integração com portais externos (ZAP, VivaReal).

--------------------------------------------------------------------------------
2. STACK TECNOLÓGICA
--------------------------------------------------------------------------------
Camada              | Tecnologia                  | Justificativa
--------------------+-----------------------------+--------------------------------
Linguagem           | Go 1.23+                    | Performance, binário único, concorrência
HTTP (back)         | Gin                         | Roteamento rápido, middlewares, ecossistema maduro
HTTP (front)        | Gin                         | SSR e roteamento web
Templates           | templ (a-h/templ)           | Tipagem forte em tempo de compilação, integra com HTMX
Interatividade      | HTMX                        | Atualização reativa de partials sem SPA complexo
CSS                 | Tailwind (CLI) ou CSS puro  | Design moderno e rápido
Banco               | PostgreSQL 16               | Relacional robusto, trigram indexes, extensões
Acesso a dados      | pgx/v5 + sqlc               | SQL puro e tipado, queries otimizadas
Migrations          | golang-migrate              | Versionamento seguro de schema (up/down)
Auth                | JWT curto + Refresh Token   | Stateless, refresh silencioso no frontend (cookies HttpOnly)
Pagamentos          | Asaas (SDK Go)              | Pix + boleto + cartão + assinaturas (Fase 3)
Upload de imagens   | S3-compatível (MinIO/R2)    | Upload direto via Presigned URLs (Browser -> Storage)
Container           | Docker + docker-compose     | Ambiente reproduzível (Postgres + MinIO)
Deploy              | Fly.io/Railway + Neon/RDS   | Baixo custo inicial e escalabilidade horizontal

--------------------------------------------------------------------------------
3. ESTRUTURA DE PASTAS
--------------------------------------------------------------------------------
imobiliaria/
├── backend/
│   ├── go.mod
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── auth/            # JWT, refresh token, hash de senha, middleware
│   │   ├── users/           # regras de usuário e gerenciamento de quota
│   │   ├── properties/      # CRUD de imóveis, filtros de busca
│   │   ├── storage/         # geração de Presigned URLs (S3/R2)
│   │   ├── payments/        # (Fase 3) Asaas integration & webhooks
│   │   ├── contracts/       # (Fase 2) contratos e assinaturas
│   │   └── db/              # pool pgx + código gerado pelo sqlc
│   ├── migrations/
│   │   ├── 0001_init.up.sql
│   │   └── 0001_init.down.sql
│   ├── sql/queries/         # arquivos .sql lidos pelo sqlc
│   ├── sqlc.yaml
│   └── Dockerfile
│
├── frontend/
│   ├── go.mod
│   ├── cmd/web/main.go
│   ├── internal/
│   │   ├── api/             # cliente HTTP tipado c/ auto-refresh de JWT
│   │   ├── handlers/        # rotas web SSR (renderizam páginas e partials)
│   │   ├── middleware/      # auth cookie extraction -> forward bearer header
│   │   └── views/           # componentes .templ
│   ├── static/
│   │   ├── htmx.min.js
│   │   └── style.css
│   └── Dockerfile
│
├── docker-compose.yml
└── README.md

--------------------------------------------------------------------------------
4. ARQUITETURA DE COMUNICAÇÃO & FLUXOS
--------------------------------------------------------------------------------
4.1 Visão Geral dos Canais
[Navegador] --HTML/HTMX--> [frontend Go :3000] --HTTP/JSON--> [backend Go :8080]
     |                              |                                  |
     |                              +-- estático (htmx, css)           +--> [PostgreSQL]
     |
     +-- Upload Direto (Presigned URL PUT) --> [MinIO / Cloudflare R2]

4.2 Fluxo de Autenticação com Auto-Refresh
1. Login/Registro: Browser faz POST no front -> Front faz POST no backend -> 
   Backend retorna `access_token` (JWT, vida de 15 min) e `refresh_token` (UUID/Hash, vida de 7 dias).
2. Frontend armazena ambos em cookies `HttpOnly`, `Secure`, `SameSite=Lax`.
3. Requisições normais: Middleware do front anexa `Authorization: Bearer <access_token>`
   nas requisições HTTP internas para a API.
4. Token Expirado (401): O cliente HTTP do frontend (`internal/api`) intercepta o 401,
   chama `POST /api/auth/refresh` enviando o refresh token, obtém um novo `access_token`,
   atualiza o cookie do usuário e reexecuta a chamada original de forma transparente.
5. Logout: Invalida o refresh token no backend e limpa os cookies.

4.3 Fluxo Otimizado de Upload de Imagens (Presigned URLs)
1. Usuário seleciona foto no painel do front.
2. Front solicita ao backend uma URL de upload: `POST /api/properties/:id/images/presign`.
3. Backend valida se o usuário é o dono do imóvel e gera uma URL assinada (S3 PUT) com expiração de 5 min.
4. Navegador faz upload direto da imagem (PUT) para o Cloudflare R2 / MinIO.
5. Após o upload concluir com 200, navegador/front confirma o registro com o backend:
   `POST /api/properties/:id/images/confirm` informando a chave/URL do arquivo e se é foto de capa.

--------------------------------------------------------------------------------
5. MODELAGEM DO BANCO DE DADOS
--------------------------------------------------------------------------------
5.1 Extensões e Tipos

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

5.2 Schema — Fase 1 (MVP)

-- ============ USERS ============
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           CITEXT UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    name            TEXT NOT NULL,
    phone           TEXT,
    cpf_cnpj        TEXT,
    role            TEXT NOT NULL DEFAULT 'USER'
                    CHECK (role IN ('USER','ADMIN')),
    plan_type       TEXT NOT NULL DEFAULT 'FREE'
                    CHECK (plan_type IN ('FREE','PAID')),
    listing_quota   INT  NOT NULL DEFAULT 2,
    listings_used   INT  NOT NULL DEFAULT 0, -- Sincronizado via trigger
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users(email);

-- ============ REFRESH TOKENS ============
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);

-- ============ PROPERTIES ============
CREATE TABLE properties (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT,
    type            TEXT NOT NULL
                    CHECK (type IN ('HOUSE','APARTMENT','LAND','COMMERCIAL')),
    purpose         TEXT NOT NULL
                    CHECK (purpose IN ('SALE','RENT')),
    status          TEXT NOT NULL DEFAULT 'ACTIVE'
                    CHECK (status IN ('ACTIVE','PAUSED','SOLD','RENTED')),
    price           NUMERIC(12,2) NOT NULL,
    condo_fee       NUMERIC(12,2),
    iptu            NUMERIC(12,2),
    area_m2         NUMERIC(8,2),
    bedrooms        SMALLINT,
    bathrooms       SMALLINT,
    parking_spots   SMALLINT,
    zip_code        VARCHAR(8), -- Apenas os 8 dígitos numéricos
    street          TEXT,
    number          TEXT,
    complement      TEXT,
    neighborhood    TEXT,
    city            TEXT NOT NULL,
    state           CHAR(2) NOT NULL,
    latitude        NUMERIC(10,8),
    longitude       NUMERIC(11,8),
    cover_image_url TEXT, -- URL desnormalizada para busca ultra-rápida sem JOIN
    views_count     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Índices otimizados para busca e performance
CREATE INDEX idx_properties_search
    ON properties (status, purpose, type, city, state, price);
CREATE INDEX idx_properties_user ON properties(user_id);
CREATE INDEX idx_properties_title_trgm
    ON properties USING gin (title gin_trgm_ops);
CREATE INDEX idx_properties_location 
    ON properties (latitude, longitude) WHERE latitude IS NOT NULL;

-- ============ PROPERTY IMAGES ============
CREATE TABLE property_images (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    position    SMALLINT NOT NULL DEFAULT 0,
    is_cover    BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_property_images_property
    ON property_images(property_id, position);
CREATE UNIQUE INDEX idx_property_single_cover
    ON property_images(property_id) WHERE is_cover = true;

-- ============ TRIGGER DE COTA DE ANÚNCIOS (INTEGRIDADE GARANTIDA) ============
CREATE OR REPLACE FUNCTION fn_sync_user_listings_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' AND NEW.status = 'ACTIVE') THEN
        UPDATE users SET listings_used = listings_used + 1, updated_at = now() WHERE id = NEW.user_id;
    ELSIF (TG_OP = 'DELETE' AND OLD.status = 'ACTIVE') THEN
        UPDATE users SET listings_used = GREATEST(0, listings_used - 1), updated_at = now() WHERE id = OLD.user_id;
    ELSIF (TG_OP = 'UPDATE') THEN
        IF (OLD.status = 'ACTIVE' AND NEW.status != 'ACTIVE') THEN
            UPDATE users SET listings_used = GREATEST(0, listings_used - 1), updated_at = now() WHERE id = NEW.user_id;
        ELSIF (OLD.status != 'ACTIVE' AND NEW.status = 'ACTIVE') THEN
            UPDATE users SET listings_used = listings_used + 1, updated_at = now() WHERE id = NEW.user_id;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sync_user_listings
AFTER INSERT OR UPDATE OF status OR DELETE ON properties
FOR EACH ROW EXECUTE FUNCTION fn_sync_user_listings_count();

-- ============ LEADS ============
CREATE TABLE leads (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id     UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    email           TEXT,
    phone           TEXT,
    message         TEXT,
    status          TEXT NOT NULL DEFAULT 'NEW'
                    CHECK (status IN ('NEW','CONTACTED','CLOSED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_leads_property ON leads(property_id);

-- ============ FAVORITES ============
CREATE TABLE favorites (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, property_id)
);

5.3 Schema — Fase 2 (contratos)

CREATE TABLE contracts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id     UUID NOT NULL REFERENCES properties(id),
    landlord_id     UUID NOT NULL REFERENCES users(id),
    tenant_id       UUID NOT NULL REFERENCES users(id),
    kind            TEXT NOT NULL CHECK (kind IN ('RENT','SALE')),
    amount          NUMERIC(12,2) NOT NULL,
    starts_at       DATE NOT NULL,
    ends_at         DATE,
    status          TEXT NOT NULL DEFAULT 'DRAFT'
                    CHECK (status IN ('DRAFT','SIGNED','ACTIVE','CLOSED')),
    pdf_url         TEXT,
    signature_ref   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

5.4 Schema — Fase 3 (pagamentos)

CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id     UUID REFERENCES contracts(id) ON DELETE SET NULL,
    user_id         UUID NOT NULL REFERENCES users(id),
    kind            TEXT NOT NULL
                    CHECK (kind IN ('RENT','SIGNAL','LISTING_PACK')),
    amount          NUMERIC(12,2) NOT NULL,
    due_date        DATE NOT NULL,
    status          TEXT NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING','PAID','OVERDUE','CANCELED')),
    method          TEXT CHECK (method IN ('PIX','BOLETO','CARD')),
    gateway_id      TEXT,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_user ON payments(user_id, status);
CREATE INDEX idx_payments_gw   ON payments(gateway_id);

--------------------------------------------------------------------------------
6. CONTRATO DA API (FASE 1)
--------------------------------------------------------------------------------
6.1 Auth
POST /api/auth/register       {name,email,password,phone} -> {id,access_token,refresh_token}
POST /api/auth/login          {email,password}            -> {user,access_token,refresh_token}
POST /api/auth/refresh        {refresh_token}             -> {access_token,refresh_token}
POST /api/auth/logout         [JWT]                       -> {message}
GET  /api/me                  [JWT]                       -> {user,quota:{used,limit}}

6.2 Imóveis
GET    /api/properties?city=&purpose=&type=&min_price=&max_price=&bedrooms=&page=&page_size=
GET    /api/properties/:id
POST   /api/properties            [JWT]  (valida quota)
PATCH  /api/properties/:id        [JWT]  (dono)
DELETE /api/properties/:id        [JWT]  (dono)

6.3 Imagens (Direct Storage Upload)
POST   /api/properties/:id/images/presign [JWT] {filename, mime_type} -> {upload_url, file_url}
POST   /api/properties/:id/images/confirm [JWT] {url, position, is_cover} -> {id, url}
DELETE /api/properties/:id/images/:img_id [JWT] (dono)

6.4 Leads
POST /api/properties/:id/leads   (público)
GET  /api/me/leads               [JWT] (dono)

6.5 Exemplo de payload (GET /api/properties)
{
  "data": [
    {
      "id": "e4b52b2f-8705-4c07-b6f3-138d6dfcecb9",
      "title": "Apartamento 2 quartos em Pinheiros",
      "type": "APARTMENT",
      "purpose": "RENT",
      "price": 3200.00,
      "condo_fee": 650.00,
      "city": "São Paulo",
      "state": "SP",
      "neighborhood": "Pinheiros",
      "bedrooms": 2,
      "bathrooms": 1,
      "parking_spots": 1,
      "area_m2": 65.5,
      "cover_image": "https://r2.exemplo.com/photos/img1.webp"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 137
}

6.6 Regra de Validação de Quota (Backend Service)
Ao criar ou ativar anúncio:
if user.listings_used >= user.listing_quota:
    return 403 { "error": "quota_exceeded",
                 "message": "Você atingiu o limite de anúncios ativos do seu plano. Faça upgrade para continuar." }

--------------------------------------------------------------------------------
7. FASES DE DESENVOLVIMENTO
--------------------------------------------------------------------------------
FASE 1 — Vitrine + Anúncios (MVP)
Objetivo: usuário se cadastra, publica imóvel, visitante busca e entra em contato.
Entregáveis:
- Backend: auth com JWT + refresh tokens, CRUD de properties, Presigned URLs para S3/MinIO,
  leads, sincronização automática de quota por trigger.
- Frontend: home com busca reativa (HTMX), detalhe do imóvel, cadastro/login,
  painel "Meus imóveis", formulário de anúncio com upload direto.
- Infra: docker-compose com PostgreSQL 16 + MinIO; migrations versionadas.
- Deploy: staging funcional (Fly.io ou Render + Neon + Cloudflare R2).
Duração: 4–5 semanas.
Critério de aceite: usuário real se cadastra, publica 2 imóveis com fotos, visitante
anônimo encontra pela busca com filtros e envia contato.

FASE 2 — Contratos
Objetivo: gerar contrato a partir de um imóvel e coletar assinatura eletrônica.
Entregáveis:
- Templates de contrato (locação e venda) parametrizados.
- Integração com Autentique ou Clicksign.
- Armazenamento do PDF assinado (S3/R2).
- Fluxo no front: gerar -> revisão -> envio -> status.
- Tabela contracts ativa.
Duração: 3–4 semanas.

FASE 3 — Pagamentos (aluguel + pacotes de anúncio)
Objetivo: cobrar aluguel recorrente e monetizar anúncios.
Entregáveis:
- Integração Asaas: cliente, cobrança Pix/boleto, assinatura recorrente.
- Webhook Asaas -> atualiza payments + incrementa listing_quota ou baixa aluguel.
- Front: área "Financeiro", botão "Comprar mais anúncios".
- Tabela payments ativa.
Duração: 4–5 semanas.
Obs.: compra de imóvel não passa por checkout; apenas sinal/reserva.
Escritura em cartório fora do escopo.

FASE 4 — Escala e Qualidade
- Notificações (e-mail via Resend + WhatsApp via Evolution API / Z-API).
- Avaliações e reputação de anunciantes.
- Moderação de anúncios (fila + admin).
- SEO avançado (sitemap dinâmico, Open Graph com imagem do imóvel, Schema.org RealEstateListing).
- Observabilidade: slog com trace_id / request_id, métricas Prometheus.
- Testes de carga (k6) na busca com 100+ req/s.

--------------------------------------------------------------------------------
8. PADRÕES E CONVENÇÕES DE CÓDIGO
--------------------------------------------------------------------------------
Backend:
- Camadas: handler (HTTP/Gin) -> service (regras de negócio e auth) -> repository (sqlc/pgx).
- Erros: pacote internal/errs com sentinels tipados (ErrNotFound, ErrQuotaExceeded, ErrUnauthorized).
- Validação: go-playground/validator nos structs DTO de entrada.
- Context: todo handler recebe ctx e propaga até o pool pgx.
- Logs: log/slog estruturado em formato JSON com request_id injetado por middleware.
- Config: variáveis de ambiente tipadas via caarlos0/env ou internal/config.

Frontend:
- Handlers apenas orquestram: chamam internal/api e renderizam views templ.
- internal/api é o cliente HTTP tipado responsável pelo auto-refresh silencioso de tokens.
- templ para componentes tipados; partials específicos para respostas HTMX.
- Middleware de auth lê cookie de sessão e injeta Authorization nas chamadas internas.

Banco:
- Toda tabela contém created_at e updated_at (onde aplicável).
- UUID como PK padrão gerado no banco (gen_random_uuid()).
- NUMERIC para valores financeiros e áreas (nunca float).
- Migrations estritas com golang-migrate (arquivos .up.sql e .down.sql).
- sqlc configurado com emit_json_tags: true e emit_interface: true.

--------------------------------------------------------------------------------
9. INFRAESTRUTURA E DEPLOY
--------------------------------------------------------------------------------
Desenvolvimento local:
  docker compose up -d db minio
  cd backend && go run ./cmd/api
  cd frontend && go run ./cmd/web

Produção (sugestão inicial econômica e escalável):
  Backend        -> Fly.io ou Railway (1 instância, auto-stop se idle)
  Frontend       -> Fly.io ou Railway (1 instância SSR)
  PostgreSQL     -> Neon Serverless ou RDS PostgreSQL 16
  Storage        -> Cloudflare R2 (zero custo de egress para imagens)
  CDN            -> Cloudflare com cache de assets estáticos e imagens

Variáveis de ambiente — Backend:
  PORT=8080
  DATABASE_URL=postgres://user:pass@host:5432/imovxen?sslmode=require
  JWT_SECRET=super-secret-key-32-chars
  JWT_TTL_MINUTES=15
  REFRESH_TOKEN_TTL_DAYS=7
  S3_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
  S3_ACCESS_KEY=...
  S3_SECRET_KEY=...
  S3_BUCKET=imovxen-images
  S3_PUBLIC_BASE_URL=https://images.imovxen.com.br
  ASAAS_API_KEY=     # Fase 3
  ASAAS_ENV=sandbox

Variáveis de ambiente — Frontend:
  PORT=3000
  BACKEND_URL=http://backend:8080
  SESSION_COOKIE_SECRET=chave-secreta-cookies-32-chars
  ENVIRONMENT=production

--------------------------------------------------------------------------------
10. SEGURANÇA
--------------------------------------------------------------------------------
- Senhas: bcrypt (cost 12) ou argon2id.
- Tokens: JWT curto (15 min) + Refresh Token rotacionado no banco, armazenado
  em cookie com flags HttpOnly, Secure e SameSite=Lax.
- Rate limiting por IP nos endpoints públicos (/api/auth/*, /api/properties, /leads).
- Upload: geração de Presigned URL restrita por extensão/MIME permitidos
  (image/jpeg, image/png, image/webp) e tamanho máximo (5 MB por foto).
- CORS no backend restrito explicitamente à URL do frontend.
- Queries protegidas contra SQL Injection via sqlc.
- Webhooks (Fase 3): verificação de assinatura criptográfica / access token no header do Asaas.

--------------------------------------------------------------------------------
11. TESTES
--------------------------------------------------------------------------------
Tipo              | Ferramenta                     | Cobertura mínima
------------------+--------------------------------+-----------------------------
Unitário (back)   | testing + testify              | Regras de quota, validações, auth
Integração (back) | testcontainers-go (Postgres)   | Queries sqlc + handlers com banco real
E2E (front)       | playwright-go                  | Cadastro -> login -> anúncio -> busca
Carga             | k6                             | Endpoint de busca suportando 100+ rps

--------------------------------------------------------------------------------
12. ROADMAP RESUMIDO
--------------------------------------------------------------------------------
Fase 1 — Vitrine + MVP  | 4–5 sem | Cadastro, anúncio c/ upload direto, busca, leads, quota
Fase 2 — Contratos      | 3–4 sem | Geração de PDFs e assinatura eletrônica
Fase 3 — Pagamentos     | 4–5 sem | Cobrança de aluguel e pacotes de anúncios (Asaas)
Fase 4 — Escala         | contínuo| WhatsApp, moderação, SEO estruturado, observabilidade

Total estimado até o fim da Fase 3: ~3 meses de desenvolvimento focado.

--------------------------------------------------------------------------------
13. PRÓXIMOS PASSOS IMEDIATOS
--------------------------------------------------------------------------------
1. Criar docker-compose.yml com Postgres 16 (extensões) + MinIO.
2. Criar backend/migrations/0001_init.up.sql e 0001_init.down.sql.
3. Configurar sqlc.yaml e escrever queries de auth e properties em backend/sql/queries/.
4. Inicializar backend/cmd/api/main.go (Gin + conexão pgx + rotas de health/properties).
5. Inicializar frontend/cmd/web/main.go com setup do templ e HTMX.
6. Configurar repositório Git com CI inicial (lint e testes).

================================================================================
FIM DO DOCUMENTO
================================================================================
