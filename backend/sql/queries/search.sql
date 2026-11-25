-- name: SearchProperties :many
SELECT id, user_id, title, description, type, purpose, status, 
       price, condo_fee, iptu, area_m2, bedrooms, bathrooms, 
       parking_spots, zip_code, street, number, complement, 
       neighborhood, city, state, latitude, longitude, cover_image_url, 
       views_count, created_at, updated_at
FROM properties
WHERE status = 'ACTIVE'
  AND ($1::text = '' OR (city ILIKE '%' || $1 || '%' OR neighborhood ILIKE '%' || $1 || '%' OR title ILIKE '%' || $1 || '%'))
  AND ($2::text = '' OR type = $2)
  AND ($3::text = '' OR purpose = $3)
  AND ($4::numeric = 0 OR price <= $4)
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;
