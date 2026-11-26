-- name: CreateContract :one
INSERT INTO contracts (
    property_id, landlord_id, tenant_name, tenant_email, tenant_phone, tenant_cpf, rent_amount, due_day, duration_months, starts_at, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetContractByID :one
SELECT c.*, p.title as property_title, p.city as property_city, p.state as property_state, p.street as property_street, p.number as property_number
FROM contracts c
JOIN properties p ON p.id = c.property_id
WHERE c.id = $1 LIMIT 1;

-- name: ListContractsByLandlord :many
SELECT c.*, p.title as property_title, p.city as property_city, p.state as property_state
FROM contracts c
JOIN properties p ON p.id = c.property_id
WHERE c.landlord_id = $1
ORDER BY c.created_at DESC;

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
SELECT i.*, c.rent_amount, c.tenant_name, c.landlord_id, p.title as property_title
FROM contract_installments i
JOIN contracts c ON c.id = i.contract_id
JOIN properties p ON p.id = c.property_id
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
