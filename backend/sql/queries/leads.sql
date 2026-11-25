-- name: CreateLead :one
INSERT INTO leads (
    property_id, name, email, phone, message
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: ListLeadsByPropertyOwner :many
SELECT l.*, p.title as property_title
FROM leads l
JOIN properties p ON p.id = l.property_id
WHERE p.user_id = $1
ORDER BY l.created_at DESC;

-- name: UpdateLeadStatus :one
UPDATE leads
SET status = $2
WHERE id = $1
RETURNING *;
