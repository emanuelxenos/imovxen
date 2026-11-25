package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Proposal struct {
	ID          uuid.UUID `json:"id"`
	PropertyID  uuid.UUID `json:"property_id"`
	UserID      uuid.UUID `json:"user_id"`
	Kind        string    `json:"kind"`
	OfferedPrice float64  `json:"offered_price"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

const listLeadsForOwner = `-- name: ListLeadsForOwner :many
SELECT l.id, l.property_id, l.name, l.email, l.phone, l.message, l.status, l.created_at, p.title as property_title
FROM leads l
JOIN properties p ON p.id = l.property_id
WHERE p.user_id = $1
ORDER BY l.created_at DESC
`

type LeadWithOwner struct {
	ID            uuid.UUID `json:"id"`
	PropertyID    uuid.UUID `json:"property_id"`
	Name          string    `json:"name"`
	Email         *string   `json:"email"`
	Phone         *string   `json:"phone"`
	Message       *string   `json:"message"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	PropertyTitle string    `json:"property_title"`
}

func (q *Queries) ListLeadsForOwner(ctx context.Context, ownerID uuid.UUID) ([]LeadWithOwner, error) {
	rows, err := q.db.Query(ctx, listLeadsForOwner, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []LeadWithOwner
	for rows.Next() {
		var i LeadWithOwner
		if err := rows.Scan(
			&i.ID,
			&i.PropertyID,
			&i.Name,
			&i.Email,
			&i.Phone,
			&i.Message,
			&i.Status,
			&i.CreatedAt,
			&i.PropertyTitle,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}
