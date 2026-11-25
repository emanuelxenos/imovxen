-- name: CreateProperty :one
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
) RETURNING *;

-- name: GetPropertyByID :one
SELECT * FROM properties
WHERE id = $1 LIMIT 1;

-- name: ListPropertiesByUser :many
SELECT * FROM properties
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdatePropertyCoverImage :exec
UPDATE properties
SET cover_image_url = $2, updated_at = now()
WHERE id = $1;

-- name: UpdatePropertyStatus :one
UPDATE properties
SET status = $2, updated_at = now()
WHERE id = $1 AND user_id = $3
RETURNING *;

-- name: DeleteProperty :exec
DELETE FROM properties
WHERE id = $1 AND user_id = $2;

-- name: IncrementPropertyViews :exec
UPDATE properties
SET views_count = views_count + 1
WHERE id = $1;

-- name: CreatePropertyImage :one
INSERT INTO property_images (
    property_id, url, position, is_cover
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: ListPropertyImages :many
SELECT * FROM property_images
WHERE property_id = $1
ORDER BY position ASC, created_at ASC;

-- name: DeletePropertyImage :exec
DELETE FROM property_images
WHERE id = $1 AND property_id = $2;

-- name: ClearPropertyCoverImages :exec
UPDATE property_images
SET is_cover = false
WHERE property_id = $1;

-- name: SetPropertyCoverImage :exec
UPDATE property_images
SET is_cover = true
WHERE id = $1 AND property_id = $2;
