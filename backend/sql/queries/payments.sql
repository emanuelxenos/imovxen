-- name: CreatePayment :one
INSERT INTO payments (
    user_id, kind, amount, items_count, due_date, status, method, gateway_id, pix_code, pix_qr_base64
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetPaymentByID :one
SELECT * FROM payments WHERE id = $1 LIMIT 1;

-- name: GetPaymentByGatewayID :one
SELECT * FROM payments WHERE gateway_id = $1 LIMIT 1;

-- name: ListPaymentsByUser :many
SELECT * FROM payments 
WHERE user_id = $1 
ORDER BY created_at DESC;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2,
    paid_at = CASE WHEN $2 = 'PAID' THEN now() ELSE paid_at END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentStatusByGatewayID :one
UPDATE payments
SET status = $2,
    paid_at = CASE WHEN $2 = 'PAID' THEN now() ELSE paid_at END,
    updated_at = now()
WHERE gateway_id = $1
RETURNING *;
