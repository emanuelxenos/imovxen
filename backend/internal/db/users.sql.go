package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, name, phone, cpf_cnpj, role, plan_type, listing_quota
) VALUES (
    $1, $2, $3, $4, $5, coalesce($6, 'USER'), coalesce($7, 'FREE'), coalesce($8, 2)
) RETURNING id, email, password_hash, name, phone, cpf_cnpj, role, plan_type, listing_quota, listings_used, created_at, updated_at
`

type CreateUserParams struct {
	Email        string  `json:"email"`
	PasswordHash string  `json:"password_hash"`
	Name         string  `json:"name"`
	Phone        *string `json:"phone"`
	CpfCnpj      *string `json:"cpf_cnpj"`
	Role         string  `json:"role"`
	PlanType     string  `json:"plan_type"`
	ListingQuota int32   `json:"listing_quota"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Email,
		arg.PasswordHash,
		arg.Name,
		arg.Phone,
		arg.CpfCnpj,
		arg.Role,
		arg.PlanType,
		arg.ListingQuota,
	)
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

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, phone, cpf_cnpj, role, plan_type, listing_quota, listings_used, created_at, updated_at 
FROM users
WHERE email = $1 LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
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

const getUserByID = `-- name: GetUserByID :one
SELECT id, email, password_hash, name, phone, cpf_cnpj, role, plan_type, listing_quota, listings_used, created_at, updated_at 
FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
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

const createRefreshToken = `-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id, token_hash, expires_at
) VALUES (
    $1, $2, $3
) RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at
`

type CreateRefreshTokenParams struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRow(ctx, createRefreshToken, arg.UserID, arg.TokenHash, arg.ExpiresAt)
	var i RefreshToken
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.TokenHash,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.CreatedAt,
	)
	return i, err
}

const getRefreshToken = `-- name: GetRefreshToken :one
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at 
FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
LIMIT 1
`

func (q *Queries) GetRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, error) {
	row := q.db.QueryRow(ctx, getRefreshToken, tokenHash)
	var i RefreshToken
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.TokenHash,
		&i.ExpiresAt,
		&i.RevokedAt,
		&i.CreatedAt,
	)
	return i, err
}

const revokeRefreshToken = `-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = $1
`

func (q *Queries) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := q.db.Exec(ctx, revokeRefreshToken, tokenHash)
	return err
}
