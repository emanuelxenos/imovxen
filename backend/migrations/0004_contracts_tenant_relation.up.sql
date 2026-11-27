-- ============ ATUALIZAÇÃO RELACIONAL ESTRITA DE CONTRATOS ============
-- O inquilino agora é OBRIGATORIAMENTE um usuário cadastrado no sistema (users.id)

ALTER TABLE contracts DROP COLUMN IF EXISTS tenant_name;
ALTER TABLE contracts DROP COLUMN IF EXISTS tenant_email;
ALTER TABLE contracts DROP COLUMN IF EXISTS tenant_phone;
ALTER TABLE contracts DROP COLUMN IF EXISTS tenant_cpf;

DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='contracts' AND column_name='tenant_id'
    ) THEN
        ALTER TABLE contracts ADD COLUMN tenant_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='contracts' AND column_name='ended_at'
    ) THEN
        ALTER TABLE contracts ADD COLUMN ended_at DATE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_contracts_tenant ON contracts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_contracts_property_history ON contracts(property_id, starts_at DESC);
