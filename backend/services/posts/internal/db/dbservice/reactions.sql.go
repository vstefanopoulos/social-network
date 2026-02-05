package dbservice

import (
	"context"
)

const getWhoLikedEntityId = `-- name: GetWhoLikedEntityId :many
SELECT user_id
FROM reactions
WHERE content_id = $1 AND deleted_at IS NULL
`

func (q *Queries) GetWhoLikedEntityId(ctx context.Context, contentID int64) ([]int64, error) {
	rows, err := q.db.Query(ctx, getWhoLikedEntityId, contentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []int64{}
	for rows.Next() {
		var user_id int64
		if err := rows.Scan(&user_id); err != nil {
			return nil, err
		}
		items = append(items, user_id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const toggleOrInsertReaction = `-- name: ToggleReaction :one
WITH toggled_off AS (
    UPDATE reactions
    SET deleted_at = NOW(),
        updated_at = NOW()
    WHERE content_id = $1
      AND user_id = $2
      AND deleted_at IS NULL
    RETURNING 1
),
restored AS (
    UPDATE reactions
    SET deleted_at = NULL,
        updated_at = NOW()
    WHERE content_id = $1
      AND user_id = $2
      AND deleted_at IS NOT NULL
      AND NOT EXISTS (SELECT 1 FROM toggled_off)
    RETURNING 1
),
inserted AS (
    INSERT INTO reactions (content_id, user_id, created_at, updated_at, deleted_at)
    SELECT $1, $2, NOW(), NOW(), NULL
    WHERE NOT EXISTS (SELECT 1 FROM toggled_off)
      AND NOT EXISTS (SELECT 1 FROM restored)
      AND NOT EXISTS (
        SELECT 1 FROM reactions WHERE content_id = $1 AND user_id = $2
      )
    RETURNING 1
)
SELECT
    CASE
        WHEN EXISTS (SELECT 1 FROM toggled_off) THEN 'removed'
        WHEN EXISTS (SELECT 1 FROM restored)   THEN 'restored'
        WHEN EXISTS (SELECT 1 FROM inserted)   THEN 'added'
        ELSE 'noop'
    END AS action,
    EXISTS (SELECT 1 FROM inserted) AS should_notify;
`

type ToggleOrInsertReactionParams struct {
	ContentID int64
	UserID    int64
}

type ToggleOrInsertReactionResult struct {
	Action       string
	ShouldNotify bool
}

func (q *Queries) ToggleOrInsertReaction(ctx context.Context, arg ToggleOrInsertReactionParams) (ToggleOrInsertReactionResult, error) {
	var res ToggleOrInsertReactionResult
	err := q.db.QueryRow(ctx, toggleOrInsertReaction, arg.ContentID, arg.UserID).
		Scan(&res.Action, &res.ShouldNotify)
	if err != nil {
		return ToggleOrInsertReactionResult{}, err
	}
	return res, nil
}
