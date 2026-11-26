-- ============ PAYMENTS ============
CREATE TABLE IF NOT EXISTS payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind            TEXT NOT NULL CHECK (kind IN ('RENT','SIGNAL','LISTING_PACK')),
    amount          NUMERIC(12,2) NOT NULL,
    items_count     INT NOT NULL DEFAULT 0, -- quantidade de anúncios adicionados caso seja LISTING_PACK
    due_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    status          TEXT NOT NULL DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING','PAID','OVERDUE','CANCELED')),
    method          TEXT NOT NULL DEFAULT 'PIX' CHECK (method IN ('PIX','BOLETO','CARD')),
    gateway_id      TEXT,
    pix_code        TEXT,
    pix_qr_base64   TEXT,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payments_user ON payments(user_id, status);
CREATE INDEX IF NOT EXISTS idx_payments_gw   ON payments(gateway_id);

-- Trigger de liberação atômica de cota quando o pagamento for concluído
CREATE OR REPLACE FUNCTION fn_process_payment_confirmation()
RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.status = 'PAID' AND (OLD.status IS NULL OR OLD.status != 'PAID')) THEN
        IF (NEW.kind = 'LISTING_PACK' AND NEW.items_count > 0) THEN
            UPDATE users 
            SET listing_quota = listing_quota + NEW.items_count,
                updated_at = now()
            WHERE id = NEW.user_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_process_payment_confirmation ON payments;
CREATE TRIGGER trg_process_payment_confirmation
AFTER UPDATE OF status ON payments
FOR EACH ROW EXECUTE FUNCTION fn_process_payment_confirmation();
