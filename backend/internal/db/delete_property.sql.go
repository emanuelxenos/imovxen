package db

import (
	"context"

	"github.com/google/uuid"
)

const deleteProperty = `-- name: DeleteProperty :exec
DELETE FROM properties
WHERE id = $1 AND user_id = $2
`

type DeletePropertyParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func (q *Queries) DeleteProperty(ctx context.Context, arg DeletePropertyParams) error {
	_, err := q.db.Exec(ctx, deleteProperty, arg.ID, arg.UserID)
	return err
}
