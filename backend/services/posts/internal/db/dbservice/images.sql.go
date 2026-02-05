package dbservice

import (
	"context"
)

const deleteImage = `-- name: DeleteImage :execrows
UPDATE images
SET deleted_at = CURRENT_TIMESTAMP
WHERE parent_id = $1 AND deleted_at IS NULL
`

// soft-deletes image with given parent_id as long as it wasn't already marked as deleted
// returns rows affected
// 0 rows affected could mean no row was found or the image was already deleted
func (q *Queries) DeleteImage(ctx context.Context, id int64) (int64, error) {
	result, err := q.db.Exec(ctx, deleteImage, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getImages = `-- name: GetImages :one
SELECT id
FROM images
WHERE parent_id = $1
  AND deleted_at IS NULL
ORDER BY sort_order
  LIMIT 1
`

// returns the first not-deleted image associated with the given parent-id
// returns norows and id 0 if no rows found
func (q *Queries) GetImages(ctx context.Context, parentID int64) (int64, error) {
	row := q.db.QueryRow(ctx, getImages, parentID)
	var id int64
	err := row.Scan(&id)
	return id, err
}

// hardcoded sort order =1 for now that we only have one image per entity id
const upsertImage = `-- name: UpsertImage :exec
INSERT INTO images (id, parent_id, sort_order)
VALUES ($1, $2, 1)
ON CONFLICT (parent_id)
DO UPDATE
SET
    id   = EXCLUDED.id,
    updated_at = CURRENT_TIMESTAMP,
    deleted_at = NULL;
`

type UpsertImageParams struct {
	ID       int64
	ParentID int64
}

// inserts a new image or updates the image associated with the parent_id if it exists
func (q *Queries) UpsertImage(ctx context.Context, arg UpsertImageParams) error {
	_, err := q.db.Exec(ctx, upsertImage, arg.ID, arg.ParentID)
	return err
}

const removeImages = `-- name: RemoveImages :exec
UPDATE images
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ANY($1::bigint[]) AND deleted_at IS NULL;
`

func (q *Queries) RemoveImages(ctx context.Context, arg []int64) error {
	_, err := q.db.Exec(ctx, removeImages, arg)
	return err
}
