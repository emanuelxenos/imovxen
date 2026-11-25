-- name: UpdateProperty :one
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
RETURNING *;
