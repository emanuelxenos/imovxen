package db

import (
	"context"
)

const searchPropertiesGeo = `-- name: SearchPropertiesGeo :many
SELECT id, user_id, title, description, type, purpose, status, 
       price, condo_fee, iptu, area_m2, bedrooms, bathrooms, 
       parking_spots, zip_code, street, number, complement, 
       neighborhood, city, state, latitude, longitude, cover_image_url, 
       views_count, created_at, updated_at,
       CASE 
         WHEN $7::numeric != 0 AND $8::numeric != 0 AND latitude IS NOT NULL AND longitude IS NOT NULL THEN
           (6371 * acos(
             least(1.0, greatest(-1.0, 
               cos(radians($7::numeric)) * cos(radians(latitude)) *
               cos(radians(longitude) - radians($8::numeric)) +
               sin(radians($7::numeric)) * sin(radians(latitude))
             ))
           ))
         ELSE 0
       END AS distance_km
FROM properties
WHERE status = 'ACTIVE'
  AND ($1::text = '' OR (city ILIKE '%' || $1 || '%' OR neighborhood ILIKE '%' || $1 || '%' OR title ILIKE '%' || $1 || '%'))
  AND ($2::text = '' OR type = $2)
  AND ($3::text = '' OR purpose = $3)
  AND ($4::numeric = 0 OR price <= $4)
  AND (
    $7::numeric = 0 OR $8::numeric = 0 OR latitude IS NULL OR longitude IS NULL OR
    (6371 * acos(
      least(1.0, greatest(-1.0,
        cos(radians($7::numeric)) * cos(radians(latitude)) *
        cos(radians(longitude) - radians($8::numeric)) +
        sin(radians($7::numeric)) * sin(radians(latitude))
      ))
    )) <= $9::numeric
  )
ORDER BY 
  CASE WHEN $7::numeric != 0 AND $8::numeric != 0 THEN 
    (6371 * acos(
      least(1.0, greatest(-1.0,
        cos(radians($7::numeric)) * cos(radians(latitude)) *
        cos(radians(longitude) - radians($8::numeric)) +
        sin(radians($7::numeric)) * sin(radians(latitude))
      ))
    ))
  END ASC,
  created_at DESC
LIMIT $5 OFFSET $6
`

type SearchPropertiesGeoParams struct {
	Query    string  `json:"query"`
	Type     string  `json:"type"`
	Purpose  string  `json:"purpose"`
	MaxPrice float64 `json:"max_price"`
	Limit    int32   `json:"limit"`
	Offset   int32   `json:"offset"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	RadiusKm float64 `json:"radius_km"`
}

type PropertyWithDistance struct {
	Property
	DistanceKm float64 `json:"distance_km"`
}

func (q *Queries) SearchPropertiesGeo(ctx context.Context, arg SearchPropertiesGeoParams) ([]PropertyWithDistance, error) {
	if arg.RadiusKm <= 0 {
		arg.RadiusKm = 50 // Raio padrão de 50km
	}

	rows, err := q.db.Query(ctx, searchPropertiesGeo,
		arg.Query,
		arg.Type,
		arg.Purpose,
		arg.MaxPrice,
		arg.Limit,
		arg.Offset,
		arg.Lat,
		arg.Lng,
		arg.RadiusKm,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PropertyWithDistance
	for rows.Next() {
		var i PropertyWithDistance
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
			&i.DistanceKm,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}
