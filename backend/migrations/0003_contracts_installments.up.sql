-- ============ CONTRACTS ============
CREATE TABLE IF NOT EXISTS contracts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id     UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    landlord_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_name     TEXT NOT NULL,
    tenant_email    TEXT,
    tenant_phone    TEXT,
    tenant_cpf      TEXT,
    rent_amount     NUMERIC(12,2) NOT NULL,
    due_day         INT NOT NULL CHECK (due_day BETWEEN 1 AND 31),
    duration_months INT NOT NULL DEFAULT 12,
    starts_at       DATE NOT NULL DEFAULT CURRENT_DATE,
    status          TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'FINISHED', 'CANCELLED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_contracts_landlord ON contracts(landlord_id);
CREATE INDEX IF NOT EXISTS idx_contracts_property ON contracts(property_id);

-- ============ CONTRACT INSTALLMENTS (CARNÊ MENSAL) ============
CREATE TABLE IF NOT EXISTS contract_installments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id         UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    installment_number  INT NOT NULL,
    due_date            DATE NOT NULL,
    amount              NUMERIC(12,2) NOT NULL,
    status              TEXT NOT NULL DEFAULT 'PENDING' 
                        CHECK (status IN ('PENDING', 'PAID_SYSTEM', 'PAID_MANUAL', 'OVERDUE')),
    manual_fee_amount   NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    paid_at             TIMESTAMPTZ,
    payment_method      TEXT,
    notes               TEXT,
    receipt_code        TEXT,
    gateway_id          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(contract_id, installment_number)
);

CREATE INDEX IF NOT EXISTS idx_installments_contract ON contract_installments(contract_id, due_date);
CREATE INDEX IF NOT EXISTS idx_installments_status   ON contract_installments(status);
