package db

import (
	"context"

	"github.com/google/uuid"
)

const createLead = `-- name: CreateLead :one
INSERT INTO leads (
    property_id, name, email, phone, message
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, property_id, name, email, phone, message, status, created_at
`

type CreateLeadParams struct {
	PropertyID uuid.UUID `json:"property_id"`
	Name       string    `json:"name"`
	Email      *string   `json:"email"`
	Phone      *string   `json:"phone"`
	Message    *string   `json:"message"`
}

func (q *Queries) CreateLead(ctx context.Context, arg CreateLeadParams) (Lead, error) {
	row := q.db.QueryRow(ctx, createLead,
		arg.PropertyID,
		arg.Name,
		arg.Email,
		arg.Phone,
		arg.Message,
	)
	var i Lead
	err := row.Scan(
		&i.ID,
		&i.PropertyID,
		&i.Name,
		&i.Email,
		&i.Phone,
		&i.Message,
		&i.Status,
		&i.CreatedAt,
	)
	return i, err
}
