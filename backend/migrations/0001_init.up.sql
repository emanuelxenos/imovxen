-- Extensões necessárias
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

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
    listings_used   INT  NOT NULL DEFAULT 0,
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
    zip_code        VARCHAR(8),
    street          TEXT,
    number          TEXT,
    complement      TEXT,
    neighborhood    TEXT,
    city            TEXT NOT NULL,
    state           CHAR(2) NOT NULL,
    latitude        NUMERIC(10,8),
    longitude       NUMERIC(11,8),
    cover_image_url TEXT,
    views_count     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

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

-- ============ TRIGGER DE COTA DE ANÚNCIOS ============
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
