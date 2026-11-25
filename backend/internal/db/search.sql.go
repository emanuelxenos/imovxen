package db

import (
	"context"
)

const searchProperties = `-- name: SearchProperties :many
SELECT id, user_id, title, description, type, purpose, status, 
       price, condo_fee, iptu, area_m2, bedrooms, bathrooms, 
       parking_spots, zip_code, street, number, complement, 
       neighborhood, city, state, latitude, longitude, cover_image_url, 
       views_count, created_at, updated_at
FROM properties
WHERE status = 'ACTIVE'
  AND ($1::text = '' OR city ILIKE '%' || $1 || '%' OR neighborhood ILIKE '%' || $1 || '%' OR title ILIKE '%' || $1 || '%')
  AND ($2::text = '' OR type = $2)
  AND ($3::text = '' OR purpose = $3)
  AND ($4::numeric = 0 OR price <= $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6
`

type SearchPropertiesParams struct {
	Query    string  `json:"query"`
	Type     string  `json:"type"`
	Purpose  string  `json:"purpose"`
	MaxPrice float64 `json:"max_price"`
	Limit    int32   `json:"limit"`
	Offset   int32   `json:"offset"`
}

func (q *Queries) SearchProperties(ctx context.Context, arg SearchPropertiesParams) ([]Property, error) {
	rows, err := q.db.Query(ctx, searchProperties,
		arg.Query,
		arg.Type,
		arg.Purpose,
		arg.MaxPrice,
		arg.Limit,
		arg.Offset,
	)
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
