package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Contract struct {
	ID             uuid.UUID  `json:"id"`
	PropertyID     uuid.UUID  `json:"property_id"`
	LandlordID     uuid.UUID  `json:"landlord_id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	RentAmount     float64    `json:"rent_amount"`
	DueDay         int32      `json:"due_day"`
	DurationMonths int32      `json:"duration_months"`
	StartsAt       time.Time  `json:"starts_at"`
	EndedAt        *time.Time `json:"ended_at"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ContractInstallment struct {
	ID                uuid.UUID  `json:"id"`
	ContractID        uuid.UUID  `json:"contract_id"`
	InstallmentNumber int32      `json:"installment_number"`
	DueDate           time.Time  `json:"due_date"`
	Amount            float64    `json:"amount"`
	Status            string     `json:"status"`
	ManualFeeAmount   float64    `json:"manual_fee_amount"`
	PaidAt            *time.Time `json:"paid_at"`
	PaymentMethod     *string    `json:"payment_method"`
	Notes             *string    `json:"notes"`
	ReceiptCode       *string    `json:"receipt_code"`
	GatewayID         *string    `json:"gateway_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type ContractWithDetails struct {
	Contract
	PropertyTitle  string  `json:"property_title"`
	PropertyCity   string  `json:"property_city"`
	PropertyState  string  `json:"property_state"`
	PropertyStreet *string `json:"property_street"`
	PropertyNumber *string `json:"property_number"`
	TenantName     string  `json:"tenant_name"`
	TenantEmail    string  `json:"tenant_email"`
	TenantPhone    *string `json:"tenant_phone"`
	TenantCpf      *string `json:"tenant_cpf"`
	LandlordName   string  `json:"landlord_name"`
	LandlordEmail  string  `json:"landlord_email"`
	LandlordPhone  *string `json:"landlord_phone"`
}

type TenantHistoryItem struct {
	ContractID     uuid.UUID  `json:"contract_id"`
	StartsAt       time.Time  `json:"starts_at"`
	EndedAt        *time.Time `json:"ended_at"`
	DurationMonths int32      `json:"duration_months"`
	RentAmount     float64    `json:"rent_amount"`
	ContractStatus string     `json:"contract_status"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	TenantName     string     `json:"tenant_name"`
	TenantEmail    string     `json:"tenant_email"`
	TenantPhone    *string    `json:"tenant_phone"`
	TenantCpf      *string    `json:"tenant_cpf"`
}

type InstallmentDetail struct {
	ContractInstallment
	RentAmount    float64   `json:"rent_amount"`
	LandlordID    uuid.UUID `json:"landlord_id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	PropertyTitle string    `json:"property_title"`
	TenantName    string    `json:"tenant_name"`
	TenantEmail   string    `json:"tenant_email"`
	TenantPhone   *string   `json:"tenant_phone"`
	LandlordName  string    `json:"landlord_name"`
	LandlordEmail string    `json:"landlord_email"`
}

type UserSearchItem struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Phone   *string   `json:"phone"`
	CpfCnpj *string   `json:"cpf_cnpj"`
	Role    string    `json:"role"`
}

const searchUsersByEmailOrName = `-- name: SearchUsersByEmailOrName :many
SELECT id, name, email, phone, cpf_cnpj, role
FROM users
WHERE (email ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')
ORDER BY name ASC
LIMIT 15
`

func (q *Queries) SearchUsersByEmailOrName(ctx context.Context, query string) ([]UserSearchItem, error) {
	rows, err := q.db.Query(ctx, searchUsersByEmailOrName, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []UserSearchItem
	for rows.Next() {
		var i UserSearchItem
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.Phone,
			&i.CpfCnpj,
			&i.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

type CreateContractParams struct {
	PropertyID     uuid.UUID `json:"property_id"`
	LandlordID     uuid.UUID `json:"landlord_id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	RentAmount     float64   `json:"rent_amount"`
	DueDay         int32     `json:"due_day"`
	DurationMonths int32     `json:"duration_months"`
	StartsAt       time.Time `json:"starts_at"`
	Status         string    `json:"status"`
}

const createContract = `-- name: CreateContract :one
INSERT INTO contracts (
    property_id, landlord_id, tenant_id, rent_amount, due_day, duration_months, starts_at, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id, property_id, landlord_id, tenant_id, rent_amount, due_day, duration_months, starts_at, ended_at, status, created_at, updated_at
`

func (q *Queries) CreateContract(ctx context.Context, arg CreateContractParams) (Contract, error) {
	row := q.db.QueryRow(ctx, createContract,
		arg.PropertyID,
		arg.LandlordID,
		arg.TenantID,
		arg.RentAmount,
		arg.DueDay,
		arg.DurationMonths,
		arg.StartsAt,
		arg.Status,
	)
	var i Contract
	err := row.Scan(
		&i.ID,
		&i.PropertyID,
		&i.LandlordID,
		&i.TenantID,
		&i.RentAmount,
		&i.DueDay,
		&i.DurationMonths,
		&i.StartsAt,
		&i.EndedAt,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getContractByID = `-- name: GetContractByID :one
SELECT c.id, c.property_id, c.landlord_id, c.tenant_id, c.rent_amount, c.due_day, c.duration_months, c.starts_at, c.ended_at, c.status, c.created_at, c.updated_at,
       p.title as property_title, p.city as property_city, p.state as property_state, p.street as property_street, p.number as property_number,
       t.name as tenant_name, t.email as tenant_email, t.phone as tenant_phone, t.cpf_cnpj as tenant_cpf,
       l.name as landlord_name, l.email as landlord_email, l.phone as landlord_phone
FROM contracts c
JOIN properties p ON p.id = c.property_id
JOIN users t ON t.id = c.tenant_id
JOIN users l ON l.id = c.landlord_id
WHERE c.id = $1 LIMIT 1
`

func (q *Queries) GetContractByID(ctx context.Context, id uuid.UUID) (ContractWithDetails, error) {
	row := q.db.QueryRow(ctx, getContractByID, id)
	var i ContractWithDetails
	err := row.Scan(
		&i.ID,
		&i.PropertyID,
		&i.LandlordID,
		&i.TenantID,
		&i.RentAmount,
		&i.DueDay,
		&i.DurationMonths,
		&i.StartsAt,
		&i.EndedAt,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.PropertyTitle,
		&i.PropertyCity,
		&i.PropertyState,
		&i.PropertyStreet,
		&i.PropertyNumber,
		&i.TenantName,
		&i.TenantEmail,
		&i.TenantPhone,
		&i.TenantCpf,
		&i.LandlordName,
		&i.LandlordEmail,
		&i.LandlordPhone,
	)
	return i, err
}

const listContractsByLandlord = `-- name: ListContractsByLandlord :many
SELECT c.id, c.property_id, c.landlord_id, c.tenant_id, c.rent_amount, c.due_day, c.duration_months, c.starts_at, c.ended_at, c.status, c.created_at, c.updated_at,
       p.title as property_title, p.city as property_city, p.state as property_state, p.street as property_street, p.number as property_number,
       t.name as tenant_name, t.email as tenant_email, t.phone as tenant_phone, t.cpf_cnpj as tenant_cpf,
       l.name as landlord_name, l.email as landlord_email, l.phone as landlord_phone
FROM contracts c
JOIN properties p ON p.id = c.property_id
JOIN users t ON t.id = c.tenant_id
JOIN users l ON l.id = c.landlord_id
WHERE c.landlord_id = $1
ORDER BY c.created_at DESC
`

func (q *Queries) ListContractsByLandlord(ctx context.Context, landlordID uuid.UUID) ([]ContractWithDetails, error) {
	rows, err := q.db.Query(ctx, listContractsByLandlord, landlordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ContractWithDetails
	for rows.Next() {
		var i ContractWithDetails
		if err := rows.Scan(
			&i.ID,
			&i.PropertyID,
			&i.LandlordID,
			&i.TenantID,
			&i.RentAmount,
			&i.DueDay,
			&i.DurationMonths,
			&i.StartsAt,
			&i.EndedAt,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.PropertyTitle,
			&i.PropertyCity,
			&i.PropertyState,
			&i.PropertyStreet,
			&i.PropertyNumber,
			&i.TenantName,
			&i.TenantEmail,
			&i.TenantPhone,
			&i.TenantCpf,
			&i.LandlordName,
			&i.LandlordEmail,
			&i.LandlordPhone,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listTenantsHistoryByProperty = `-- name: ListTenantsHistoryByProperty :many
SELECT c.id as contract_id, c.starts_at, c.ended_at, c.duration_months, c.rent_amount, c.status as contract_status,
       t.id as tenant_id, t.name as tenant_name, t.email as tenant_email, t.phone as tenant_phone, t.cpf_cnpj as tenant_cpf
FROM contracts c
JOIN users t ON t.id = c.tenant_id
WHERE c.property_id = $1
ORDER BY c.starts_at DESC
`

func (q *Queries) ListTenantsHistoryByProperty(ctx context.Context, propertyID uuid.UUID) ([]TenantHistoryItem, error) {
	rows, err := q.db.Query(ctx, listTenantsHistoryByProperty, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TenantHistoryItem
	for rows.Next() {
		var i TenantHistoryItem
		if err := rows.Scan(
			&i.ContractID,
			&i.StartsAt,
			&i.EndedAt,
			&i.DurationMonths,
			&i.RentAmount,
			&i.ContractStatus,
			&i.TenantID,
			&i.TenantName,
			&i.TenantEmail,
			&i.TenantPhone,
			&i.TenantCpf,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const finalizeContract = `-- name: FinalizeContract :one
UPDATE contracts
SET status = 'FINISHED',
    ended_at = CURRENT_DATE,
    updated_at = now()
WHERE id = $1 AND landlord_id = $2
RETURNING id, property_id, landlord_id, tenant_id, rent_amount, due_day, duration_months, starts_at, ended_at, status, created_at, updated_at
`

func (q *Queries) FinalizeContract(ctx context.Context, id uuid.UUID, landlordID uuid.UUID) (Contract, error) {
	row := q.db.QueryRow(ctx, finalizeContract, id, landlordID)
	var i Contract
	err := row.Scan(
		&i.ID,
		&i.PropertyID,
		&i.LandlordID,
		&i.TenantID,
		&i.RentAmount,
		&i.DueDay,
		&i.DurationMonths,
		&i.StartsAt,
		&i.EndedAt,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

type CreateInstallmentParams struct {
	ContractID        uuid.UUID `json:"contract_id"`
	InstallmentNumber int32     `json:"installment_number"`
	DueDate           time.Time `json:"due_date"`
	Amount            float64   `json:"amount"`
	Status            string    `json:"status"`
}

const createInstallment = `-- name: CreateInstallment :one
INSERT INTO contract_installments (
    contract_id, installment_number, due_date, amount, status
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, contract_id, installment_number, due_date, amount, status, manual_fee_amount, paid_at, payment_method, notes, receipt_code, gateway_id, created_at, updated_at
`

func (q *Queries) CreateInstallment(ctx context.Context, arg CreateInstallmentParams) (ContractInstallment, error) {
	row := q.db.QueryRow(ctx, createInstallment,
		arg.ContractID,
		arg.InstallmentNumber,
		arg.DueDate,
		arg.Amount,
		arg.Status,
	)
	var i ContractInstallment
	err := row.Scan(
		&i.ID,
		&i.ContractID,
		&i.InstallmentNumber,
		&i.DueDate,
		&i.Amount,
		&i.Status,
		&i.ManualFeeAmount,
		&i.PaidAt,
		&i.PaymentMethod,
		&i.Notes,
		&i.ReceiptCode,
		&i.GatewayID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listInstallmentsByContract = `-- name: ListInstallmentsByContract :many
SELECT id, contract_id, installment_number, due_date, amount, status, manual_fee_amount, paid_at, payment_method, notes, receipt_code, gateway_id, created_at, updated_at
FROM contract_installments
WHERE contract_id = $1
ORDER BY installment_number ASC
`

func (q *Queries) ListInstallmentsByContract(ctx context.Context, contractID uuid.UUID) ([]ContractInstallment, error) {
	rows, err := q.db.Query(ctx, listInstallmentsByContract, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ContractInstallment
	for rows.Next() {
		var i ContractInstallment
		if err := rows.Scan(
			&i.ID,
			&i.ContractID,
			&i.InstallmentNumber,
			&i.DueDate,
			&i.Amount,
			&i.Status,
			&i.ManualFeeAmount,
			&i.PaidAt,
			&i.PaymentMethod,
			&i.Notes,
			&i.ReceiptCode,
			&i.GatewayID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const getInstallmentByID = `-- name: GetInstallmentByID :one
SELECT i.id, i.contract_id, i.installment_number, i.due_date, i.amount, i.status, i.manual_fee_amount, i.paid_at, i.payment_method, i.notes, i.receipt_code, i.gateway_id, i.created_at, i.updated_at,
       c.rent_amount, c.landlord_id, c.tenant_id, p.title as property_title,
       t.name as tenant_name, t.email as tenant_email, t.phone as tenant_phone,
       l.name as landlord_name, l.email as landlord_email
FROM contract_installments i
JOIN contracts c ON c.id = i.contract_id
JOIN properties p ON p.id = c.property_id
JOIN users t ON t.id = c.tenant_id
JOIN users l ON l.id = c.landlord_id
WHERE i.id = $1 LIMIT 1
`

func (q *Queries) GetInstallmentByID(ctx context.Context, id uuid.UUID) (InstallmentDetail, error) {
	row := q.db.QueryRow(ctx, getInstallmentByID, id)
	var i InstallmentDetail
	err := row.Scan(
		&i.ID,
		&i.ContractID,
		&i.InstallmentNumber,
		&i.DueDate,
		&i.Amount,
		&i.Status,
		&i.ManualFeeAmount,
		&i.PaidAt,
		&i.PaymentMethod,
		&i.Notes,
		&i.ReceiptCode,
		&i.GatewayID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RentAmount,
		&i.LandlordID,
		&i.TenantID,
		&i.PropertyTitle,
		&i.TenantName,
		&i.TenantEmail,
		&i.TenantPhone,
		&i.LandlordName,
		&i.LandlordEmail,
	)
	return i, err
}

type MarkInstallmentPaidManualParams struct {
	ID              uuid.UUID `json:"id"`
	ManualFeeAmount float64   `json:"manual_fee_amount"`
	PaymentMethod   *string   `json:"payment_method"`
	Notes           *string   `json:"notes"`
	ReceiptCode     *string   `json:"receipt_code"`
}

const markInstallmentPaidManual = `-- name: MarkInstallmentPaidManual :one
UPDATE contract_installments
SET status = 'PAID_MANUAL',
    manual_fee_amount = $2,
    paid_at = now(),
    payment_method = $3,
    notes = $4,
    receipt_code = $5,
    updated_at = now()
WHERE id = $1
RETURNING id, contract_id, installment_number, due_date, amount, status, manual_fee_amount, paid_at, payment_method, notes, receipt_code, gateway_id, created_at, updated_at
`

func (q *Queries) MarkInstallmentPaidManual(ctx context.Context, arg MarkInstallmentPaidManualParams) (ContractInstallment, error) {
	row := q.db.QueryRow(ctx, markInstallmentPaidManual,
		arg.ID,
		arg.ManualFeeAmount,
		arg.PaymentMethod,
		arg.Notes,
		arg.ReceiptCode,
	)
	var i ContractInstallment
	err := row.Scan(
		&i.ID,
		&i.ContractID,
		&i.InstallmentNumber,
		&i.DueDate,
		&i.Amount,
		&i.Status,
		&i.ManualFeeAmount,
		&i.PaidAt,
		&i.PaymentMethod,
		&i.Notes,
		&i.ReceiptCode,
		&i.GatewayID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

type UpdateInstallmentDueDateParams struct {
	ID        uuid.UUID `json:"id"`
	DueDate   time.Time `json:"due_date"`
	GatewayID *string   `json:"gateway_id"`
}

const updateInstallmentDueDate = `-- name: UpdateInstallmentDueDate :one
UPDATE contract_installments
SET due_date = $2,
    gateway_id = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, contract_id, installment_number, due_date, amount, status, manual_fee_amount, paid_at, payment_method, notes, receipt_code, gateway_id, created_at, updated_at
`

func (q *Queries) UpdateInstallmentDueDate(ctx context.Context, arg UpdateInstallmentDueDateParams) (ContractInstallment, error) {
	row := q.db.QueryRow(ctx, updateInstallmentDueDate,
		arg.ID,
		arg.DueDate,
		arg.GatewayID,
	)
	var i ContractInstallment
	err := row.Scan(
		&i.ID,
		&i.ContractID,
		&i.InstallmentNumber,
		&i.DueDate,
		&i.Amount,
		&i.Status,
		&i.ManualFeeAmount,
		&i.PaidAt,
		&i.PaymentMethod,
		&i.Notes,
		&i.ReceiptCode,
		&i.GatewayID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
