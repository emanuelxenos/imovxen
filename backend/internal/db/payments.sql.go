package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Kind        string     `json:"kind"`
	Amount      float64    `json:"amount"`
	ItemsCount  int32      `json:"items_count"`
	DueDate     time.Time  `json:"due_date"`
	Status      string     `json:"status"`
	Method      string     `json:"method"`
	GatewayID   *string    `json:"gateway_id"`
	PixCode     *string    `json:"pix_code"`
	PixQrBase64 *string    `json:"pix_qr_base64"`
	PaidAt      *time.Time `json:"paid_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreatePaymentParams struct {
	UserID      uuid.UUID `json:"user_id"`
	Kind        string    `json:"kind"`
	Amount      float64   `json:"amount"`
	ItemsCount  int32     `json:"items_count"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
	Method      string    `json:"method"`
	GatewayID   *string   `json:"gateway_id"`
	PixCode     *string   `json:"pix_code"`
	PixQrBase64 *string   `json:"pix_qr_base64"`
}

const createPayment = `-- name: CreatePayment :one
INSERT INTO payments (
    user_id, kind, amount, items_count, due_date, status, method, gateway_id, pix_code, pix_qr_base64
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING id, user_id, kind, amount, items_count, due_date, status, method, gateway_id, pix_code, pix_qr_base64, paid_at, created_at, updated_at
`

func (q *Queries) CreatePayment(ctx context.Context, arg CreatePaymentParams) (Payment, error) {
	row := q.db.QueryRow(ctx, createPayment,
		arg.UserID,
		arg.Kind,
		arg.Amount,
		arg.ItemsCount,
		arg.DueDate,
		arg.Status,
		arg.Method,
		arg.GatewayID,
		arg.PixCode,
		arg.PixQrBase64,
	)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Kind,
		&i.Amount,
		&i.ItemsCount,
		&i.DueDate,
		&i.Status,
		&i.Method,
		&i.GatewayID,
		&i.PixCode,
		&i.PixQrBase64,
		&i.PaidAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPaymentByID = `-- name: GetPaymentByID :one
SELECT id, user_id, kind, amount, items_count, due_date, status, method, gateway_id, pix_code, pix_qr_base64, paid_at, created_at, updated_at
FROM payments WHERE id = $1 LIMIT 1
`

func (q *Queries) GetPaymentByID(ctx context.Context, id uuid.UUID) (Payment, error) {
	row := q.db.QueryRow(ctx, getPaymentByID, id)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Kind,
		&i.Amount,
		&i.ItemsCount,
		&i.DueDate,
		&i.Status,
		&i.Method,
		&i.GatewayID,
		&i.PixCode,
		&i.PixQrBase64,
		&i.PaidAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPaymentByGatewayID = `-- name: GetPaymentByGatewayID :one
SELECT id, user_id, kind, amount, items_count, due_date, status, method, gateway_id, pix_code, pix_qr_base64, paid_at, created_at, updated_at
FROM payments WHERE gateway_id = $1 LIMIT 1
`

func (q *Queries) GetPaymentByGatewayID(ctx context.Context, gatewayID string) (Payment, error) {
	row := q.db.QueryRow(ctx, getPaymentByGatewayID, gatewayID)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Kind,
		&i.Amount,
		&i.ItemsCount,
		&i.DueDate,
		&i.Status,
		&i.Method,
		&i.GatewayID,
		&i.PixCode,
		&i.PixQrBase64,
		&i.PaidAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listPaymentsByUser = `-- name: ListPaymentsByUser :many
SELECT id, user_id, kind, amount, items_count, due_date, status, method, gateway_id, pix_code, pix_qr_base64, paid_at, created_at, updated_at
FROM payments
WHERE user_id = $1
ORDER BY created_at DESC
`

func (q *Queries) ListPaymentsByUser(ctx context.Context, userID uuid.UUID) ([]Payment, error) {
	rows, err := q.db.Query(ctx, listPaymentsByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Payment
	for rows.Next() {
		var i Payment
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Kind,
			&i.Amount,
			&i.ItemsCount,
			&i.DueDate,
			&i.Status,
			&i.Method,
			&i.GatewayID,
			&i.PixCode,
			&i.PixQrBase64,
			&i.PaidAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const updatePaymentStatus = `-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2,
    paid_at = CASE WHEN $2 = 'PAID' THEN now() ELSE paid_at END,
    updated_at = now()
WHERE id = $1
RETURNING id, user_id, kind, amount, items_count, due_date, status, method, gateway_id, pix_code, pix_qr_base64, paid_at, created_at, updated_at
`

func (q *Queries) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string) (Payment, error) {
	row := q.db.QueryRow(ctx, updatePaymentStatus, id, status)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Kind,
		&i.Amount,
		&i.ItemsCount,
		&i.DueDate,
		&i.Status,
		&i.Method,
		&i.GatewayID,
		&i.PixCode,
		&i.PixQrBase64,
		&i.PaidAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updatePaymentStatusByGatewayID = `-- name: UpdatePaymentStatusByGatewayID :one
UPDATE payments
SET status = $2,
    paid_at = CASE WHEN $2 = 'PAID' THEN now() ELSE paid_at END,
    updated_at = now()
WHERE gateway_id = $1
RETURNING id, user_id, kind, amount, items_count, due_date, status, method, gateway_id, pix_code, pix_qr_base64, paid_at, created_at, updated_at
`

func (q *Queries) UpdatePaymentStatusByGatewayID(ctx context.Context, gatewayID string, status string) (Payment, error) {
	row := q.db.QueryRow(ctx, updatePaymentStatusByGatewayID, gatewayID, status)
	var i Payment
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Kind,
		&i.Amount,
		&i.ItemsCount,
		&i.DueDate,
		&i.Status,
		&i.Method,
		&i.GatewayID,
		&i.PixCode,
		&i.PixQrBase64,
		&i.PaidAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
