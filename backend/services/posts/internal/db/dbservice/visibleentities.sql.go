package dbservice

import (
	"context"
)

const canUserSeeEntity = `-- name: CanUserSeeEntity :one
WITH ent AS (
    -- Post
    SELECT 
        id,
        creator_id,
        audience,
        group_id
    FROM posts
    WHERE id = $4::bigint
      AND deleted_at IS NULL

    UNION ALL

    -- Event
    SELECT
        id,
        event_creator_id,
        NULL AS audience,
        group_id
    FROM events
    WHERE id = $4::bigint
      AND deleted_at IS NULL
)
SELECT EXISTS (
    SELECT 1
    FROM ent e
    WHERE
        (
            -- CASE 0: creator can always see
            e.creator_id = $3::bigint
        )
        OR
        (
            -- CASE 1: group entity
            e.group_id IS NOT NULL
            AND $1::bool = TRUE
        )
        OR
        (
            -- CASE 2: post (no group)
            e.group_id IS NULL
            AND (
                e.audience = 'everyone'
                OR (e.audience = 'followers' AND $2::bool = TRUE)
                OR (
                    e.audience = 'selected'
                    AND EXISTS (
                        SELECT 1 FROM post_audience pa
                        WHERE pa.post_id = e.id
                          AND pa.allowed_user_id = $3::bigint
                    )
                )
            )
        )
)
`

type CanUserSeeEntityParams struct {
	IsMember    bool
	IsFollowing bool
	UserID      int64
	EntityID    int64
}

func (q *Queries) CanUserSeeEntity(ctx context.Context, arg CanUserSeeEntityParams) (bool, error) {
	row := q.db.QueryRow(ctx, canUserSeeEntity,
		arg.IsMember,
		arg.IsFollowing,
		arg.UserID,
		arg.EntityID,
	)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const getEntityCreatorAndGroup = `-- name: GetEntityCreatorAndGroup :one
SELECT
    mi.content_type,

     -- Who created THIS content (comment author, post author, event creator)
    (
        CASE
            WHEN mi.content_type = 'post'
                THEN p.creator_id
            WHEN mi.content_type = 'event'
                THEN e.event_creator_id
            WHEN mi.content_type = 'comment'
                THEN c.comment_creator_id
        END
    )::BIGINT AS creator_id,

    -- creator of the parent post (only for comments)
    (
        CASE
            WHEN mi.content_type = 'comment'
                THEN p2.creator_id
            ELSE 0
        END
    )::BIGINT AS parent_creator_id,

    -- group_id: post.group_id, event.group_id, or parent post group for comments
COALESCE(
    CASE
        WHEN mi.content_type = 'post'    THEN p.group_id
        WHEN mi.content_type = 'event'   THEN e.group_id
        WHEN mi.content_type = 'comment' THEN p2.group_id
    END,
    0
)::BIGINT AS group_id,

-- parent post id (for comments)
    CASE
        WHEN mi.content_type = 'comment' THEN c.parent_id
        ELSE 0
    END::BIGINT AS parent_id

FROM master_index mi
LEFT JOIN posts p ON p.id = mi.id
LEFT JOIN events e ON e.id = mi.id
LEFT JOIN comments c ON c.id = mi.id
LEFT JOIN posts p2 ON p2.id = c.parent_id  -- parent post for comments
WHERE mi.id = $1
LIMIT 1
`

type GetEntityCreatorAndGroupRow struct {
	ContentType     ContentType
	CreatorID       int64
	ParentCreatorID int64
	GroupID         int64
	ParentID        int64
}

func (q *Queries) GetEntityCreatorAndGroup(ctx context.Context, id int64) (GetEntityCreatorAndGroupRow, error) {
	row := q.db.QueryRow(ctx, getEntityCreatorAndGroup, id)
	var i GetEntityCreatorAndGroupRow
	err := row.Scan(&i.ContentType, &i.CreatorID, &i.ParentCreatorID, &i.GroupID, &i.ParentID)
	return i, err
}
