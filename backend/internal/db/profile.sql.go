package db

import (
	"context"

	"github.com/google/uuid"
)

const updateUserProfile = `-- name: UpdateUserProfile :one
UPDATE users
SET name = $2,
    phone = $3,
    cpf_cnpj = $4,
    updated_at = now()
WHERE id = $1
RETURNING id, email, password_hash, name, phone, cpf_cnpj, role, plan_type, listing_quota, listings_used, created_at, updated_at
`

type UpdateUserProfileParams struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Phone   *string   `json:"phone"`
	CpfCnpj *string   `json:"cpf_cnpj"`
}

func (q *Queries) UpdateUserProfile(ctx context.Context, arg UpdateUserProfileParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserProfile, arg.ID, arg.Name, arg.Phone, arg.CpfCnpj)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.Phone,
		&i.CpfCnpj,
		&i.Role,
		&i.PlanType,
		&i.ListingQuota,
		&i.ListingsUsed,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2,
    updated_at = now()
WHERE id = $1
`

type UpdateUserPasswordParams struct {
	ID           uuid.UUID `json:"id"`
	PasswordHash string    `json:"password_hash"`
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.Exec(ctx, updateUserPassword, arg.ID, arg.PasswordHash)
	return err
}
