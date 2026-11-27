-- name: SearchUsersByEmailOrName :many
SELECT id, name, email, phone, cpf_cnpj, role
FROM users
WHERE (email ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')
ORDER BY name ASC
LIMIT 15;

-- name: CreateContract :one
INSERT INTO contracts (
    property_id, landlord_id, tenant_id, rent_amount, due_day, duration_months, starts_at, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetContractByID :one
SELECT c.*, 
       p.title as property_title, p.city as property_city, p.state as property_state, p.street as property_street, p.number as property_number,
       t.name as tenant_name, t.email as tenant_email, t.phone as tenant_phone, t.cpf_cnpj as tenant_cpf,
       l.name as landlord_name, l.email as landlord_email, l.phone as landlord_phone
FROM contracts c
JOIN properties p ON p.id = c.property_id
JOIN users t ON t.id = c.tenant_id
JOIN users l ON l.id = c.landlord_id
WHERE c.id = $1 LIMIT 1;

-- name: ListContractsByLandlord :many
SELECT c.*, 
       p.title as property_title, p.city as property_city, p.state as property_state,
       t.name as tenant_name, t.email as tenant_email, t.phone as tenant_phone, t.cpf_cnpj as tenant_cpf
FROM contracts c
JOIN properties p ON p.id = c.property_id
JOIN users t ON t.id = c.tenant_id
WHERE c.landlord_id = $1
ORDER BY c.created_at DESC;

-- name: ListTenantsHistoryByProperty :many
SELECT c.id as contract_id, c.starts_at, c.ended_at, c.duration_months, c.rent_amount, c.status as contract_status,
       t.id as tenant_id, t.name as tenant_name, t.email as tenant_email, t.phone as tenant_phone, t.cpf_cnpj as tenant_cpf
FROM contracts c
JOIN users t ON t.id = c.tenant_id
WHERE c.property_id = $1
ORDER BY c.starts_at DESC;

-- name: FinalizeContract :one
UPDATE contracts
SET status = 'FINISHED',
    ended_at = CURRENT_DATE,
    updated_at = now()
WHERE id = $1 AND landlord_id = $2
RETURNING *;

-- name: CreateInstallment :one
INSERT INTO contract_installments (
    contract_id, installment_number, due_date, amount, status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: ListInstallmentsByContract :many
SELECT * FROM contract_installments
WHERE contract_id = $1
ORDER BY installment_number ASC;

-- name: GetInstallmentByID :one
SELECT i.*, c.rent_amount, c.landlord_id, c.tenant_id, p.title as property_title,
       t.name as tenant_name, t.email as tenant_email, t.phone as tenant_phone,
       l.name as landlord_name, l.email as landlord_email
FROM contract_installments i
JOIN contracts c ON c.id = i.contract_id
JOIN properties p ON p.id = c.property_id
JOIN users t ON t.id = c.tenant_id
JOIN users l ON l.id = c.landlord_id
WHERE i.id = $1 LIMIT 1;

-- name: MarkInstallmentPaidManual :one
UPDATE contract_installments
SET status = 'PAID_MANUAL',
    manual_fee_amount = $2,
    paid_at = now(),
    payment_method = $3,
    notes = $4,
    receipt_code = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateInstallmentDueDate :one
UPDATE contract_installments
SET due_date = $2,
    gateway_id = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;
