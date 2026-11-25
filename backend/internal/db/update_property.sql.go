package db

import (
	"context"

	"github.com/google/uuid"
)

const updateProperty = `-- name: UpdateProperty :one
UPDATE properties
SET title = $2,
    description = $3,
    type = $4,
    purpose = $5,
    price = $6,
    condo_fee = $7,
    iptu = $8,
    area_m2 = $9,
    bedrooms = $10,
    bathrooms = $11,
    parking_spots = $12,
    city = $13,
    state = $14,
    cover_image_url = $15,
    updated_at = now()
WHERE id = $1 AND user_id = $16
RETURNING id, user_id, title, description, type, purpose, status, price, condo_fee, iptu, area_m2, bedrooms, bathrooms, parking_spots, zip_code, street, number, complement, neighborhood, city, state, latitude, longitude, cover_image_url, views_count, created_at, updated_at
`

type UpdatePropertyParams struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description"`
	Type          string    `json:"type"`
	Purpose       string    `json:"purpose"`
	Price         float64   `json:"price"`
	CondoFee      *float64  `json:"condo_fee"`
	Iptu          *float64  `json:"iptu"`
	AreaM2        *float64  `json:"area_m2"`
	Bedrooms      *int16    `json:"bedrooms"`
	Bathrooms     *int16    `json:"bathrooms"`
	ParkingSpots  *int16    `json:"parking_spots"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	CoverImageUrl *string   `json:"cover_image_url"`
	UserID        uuid.UUID `json:"user_id"`
}

func (q *Queries) UpdateProperty(ctx context.Context, arg UpdatePropertyParams) (Property, error) {
	row := q.db.QueryRow(ctx, updateProperty,
		arg.ID,
		arg.Title,
		arg.Description,
		arg.Type,
		arg.Purpose,
		arg.Price,
		arg.CondoFee,
		arg.Iptu,
		arg.AreaM2,
		arg.Bedrooms,
		arg.Bathrooms,
		arg.ParkingSpots,
		arg.City,
		arg.State,
		arg.CoverImageUrl,
		arg.UserID,
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
