package db

import (
	"context"

	"github.com/google/uuid"
)

const createProperty = `-- name: CreateProperty :one
INSERT INTO properties (
    user_id, title, description, type, purpose, status,
    price, condo_fee, iptu, area_m2, bedrooms, bathrooms,
    parking_spots, zip_code, street, number, complement,
    neighborhood, city, state, latitude, longitude, cover_image_url
) VALUES (
    $1, $2, $3, $4, $5, coalesce($6, 'ACTIVE'),
    $7, $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17,
    $18, $19, $20, $21, $22, $23
) RETURNING id, user_id, title, description, type, purpose, status, price, condo_fee, iptu, area_m2, bedrooms, bathrooms, parking_spots, zip_code, street, number, complement, neighborhood, city, state, latitude, longitude, cover_image_url, views_count, created_at, updated_at
`

type CreatePropertyParams struct {
	UserID        uuid.UUID `json:"user_id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description"`
	Type          string    `json:"type"`
	Purpose       string    `json:"purpose"`
	Status        string    `json:"status"`
	Price         float64   `json:"price"`
	CondoFee      *float64  `json:"condo_fee"`
	Iptu          *float64  `json:"iptu"`
	AreaM2        *float64  `json:"area_m2"`
	Bedrooms      *int16    `json:"bedrooms"`
	Bathrooms     *int16    `json:"bathrooms"`
	ParkingSpots  *int16    `json:"parking_spots"`
	ZipCode       *string   `json:"zip_code"`
	Street        *string   `json:"street"`
	Number        *string   `json:"number"`
	Complement    *string   `json:"complement"`
	Neighborhood  *string   `json:"neighborhood"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	Latitude      *float64  `json:"latitude"`
	Longitude     *float64  `json:"longitude"`
	CoverImageUrl *string   `json:"cover_image_url"`
}

func (q *Queries) CreateProperty(ctx context.Context, arg CreatePropertyParams) (Property, error) {
	row := q.db.QueryRow(ctx, createProperty,
		arg.UserID,
		arg.Title,
		arg.Description,
		arg.Type,
		arg.Purpose,
		arg.Status,
		arg.Price,
		arg.CondoFee,
		arg.Iptu,
		arg.AreaM2,
		arg.Bedrooms,
		arg.Bathrooms,
		arg.ParkingSpots,
		arg.ZipCode,
		arg.Street,
		arg.Number,
		arg.Complement,
		arg.Neighborhood,
		arg.City,
		arg.State,
		arg.Latitude,
		arg.Longitude,
		arg.CoverImageUrl,
	)
	var i Property
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Title,
		&i.Description,
		&i.Type,
		&i.Purpose,
		&i.Status,
		&i.Price,
		&i.CondoFee,
		&i.Iptu,
		&i.AreaM2,
		&i.Bedrooms,
		&i.Bathrooms,
		&i.ParkingSpots,
		&i.ZipCode,
		&i.Street,
		&i.Number,
		&i.Complement,
		&i.Neighborhood,
		&i.City,
		&i.State,
		&i.Latitude,
		&i.Longitude,
		&i.CoverImageUrl,
		&i.ViewsCount,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPropertyByID = `-- name: GetPropertyByID :one
SELECT id, user_id, title, description, type, purpose, status, price, condo_fee, iptu, area_m2, bedrooms, bathrooms, parking_spots, zip_code, street, number, complement, neighborhood, city, state, latitude, longitude, cover_image_url, views_count, created_at, updated_at 
FROM properties
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetPropertyByID(ctx context.Context, id uuid.UUID) (Property, error) {
	row := q.db.QueryRow(ctx, getPropertyByID, id)
	var i Property
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Title,
		&i.Description,
		&i.Type,
		&i.Purpose,
		&i.Status,
		&i.Price,
		&i.CondoFee,
		&i.Iptu,
		&i.AreaM2,
		&i.Bedrooms,
		&i.Bathrooms,
		&i.ParkingSpots,
		&i.ZipCode,
		&i.Street,
		&i.Number,
		&i.Complement,
		&i.Neighborhood,
		&i.City,
		&i.State,
		&i.Latitude,
		&i.Longitude,
		&i.CoverImageUrl,
		&i.ViewsCount,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listPropertiesByUser = `-- name: ListPropertiesByUser :many
SELECT id, user_id, title, description, type, purpose, status, price, condo_fee, iptu, area_m2, bedrooms, bathrooms, parking_spots, zip_code, street, number, complement, neighborhood, city, state, latitude, longitude, cover_image_url, views_count, created_at, updated_at 
FROM properties
WHERE user_id = $1
ORDER BY created_at DESC
`

func (q *Queries) ListPropertiesByUser(ctx context.Context, userID uuid.UUID) ([]Property, error) {
	rows, err := q.db.Query(ctx, listPropertiesByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Property
	for rows.Next() {
		var i Property
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Title,
			&i.Description,
			&i.Type,
			&i.Purpose,
			&i.Status,
			&i.Price,
			&i.CondoFee,
			&i.Iptu,
			&i.AreaM2,
			&i.Bedrooms,
			&i.Bathrooms,
			&i.ParkingSpots,
			&i.ZipCode,
			&i.Street,
			&i.Number,
			&i.Complement,
			&i.Neighborhood,
			&i.City,
			&i.State,
			&i.Latitude,
			&i.Longitude,
			&i.CoverImageUrl,
			&i.ViewsCount,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}
