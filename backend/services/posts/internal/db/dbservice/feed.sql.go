package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getGroupPostsPaginated = `-- name: GetGroupPostsPaginated :many
SELECT
    p.id,
    p.post_body,
    p.creator_id,
    p.group_id,
    p.audience,
    p.comments_count,
    p.reactions_count,
    p.last_commented_at,
    p.created_at,
    p.updated_at,

    EXISTS (     -- Has the given user liked the post?
        SELECT 1
        FROM reactions r
        WHERE r.content_id = p.id
          AND r.user_id = $2              -- requesting user (check is member from users service)
          AND r.deleted_at IS NULL
    ) AS liked_by_user,
   
COALESCE(
    (SELECT i.id
     FROM images i
     WHERE i.parent_id = p.id AND i.deleted_at IS NULL
     ORDER BY i.sort_order ASC
     LIMIT 1
    ), 0
)::bigint AS image  

  
FROM posts p



WHERE p.group_id = $1                    -- group id filter
  AND p.deleted_at IS NULL
GROUP BY p.id
ORDER BY p.created_at DESC               -- newest first
LIMIT $3 OFFSET $4
`

type GetGroupPostsPaginatedParams struct {
	GroupID pgtype.Int8
	UserID  int64
	Limit   int32
	Offset  int32
}

type GetGroupPostsPaginatedRow struct {
	ID              int64
	PostBody        string
	CreatorID       int64
	GroupID         pgtype.Int8
	Audience        IntendedAudience
	CommentsCount   int32
	ReactionsCount  int32
	LastCommentedAt pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	LikedByUser     bool
	Image           int64
}

func (q *Queries) GetGroupPostsPaginated(ctx context.Context, arg GetGroupPostsPaginatedParams) ([]GetGroupPostsPaginatedRow, error) {
	rows, err := q.db.Query(ctx, getGroupPostsPaginated,
		arg.GroupID,
		arg.UserID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetGroupPostsPaginatedRow{}
	for rows.Next() {
		var i GetGroupPostsPaginatedRow
		if err := rows.Scan(
			&i.ID,
			&i.PostBody,
			&i.CreatorID,
			&i.GroupID,
			&i.Audience,
			&i.CommentsCount,
			&i.ReactionsCount,
			&i.LastCommentedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LikedByUser,
			&i.Image,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPersonalizedFeed = `-- name: GetPersonalizedFeed :many
SELECT
    p.id,
    p.post_body,
    p.creator_id,
    p.comments_count,
    p.reactions_count,
    p.last_commented_at,
    p.created_at,
    p.updated_at,

    -- did user like it?
    EXISTS (
        SELECT 1 FROM reactions r
        WHERE r.content_id = p.id
          AND r.user_id = $1
          AND r.deleted_at IS NULL
    ) AS liked_by_user,

    -- image
COALESCE(
    (SELECT i.id
     FROM images i
     WHERE i.parent_id = p.id AND i.deleted_at IS NULL
     ORDER BY i.sort_order ASC
     LIMIT 1
    ), 0
)::bigint AS image   

   FROM posts p



WHERE p.deleted_at IS NULL
  AND (
       -- SELECTED audience → only manually approved viewers
       (p.audience = 'selected' AND EXISTS (
           SELECT 1 FROM post_audience pa
           WHERE pa.post_id = p.id AND pa.allowed_user_id = $1
       ))

       -- FOLLOWERS → allowed if creator ∈ list passed in
       OR (p.audience = 'followers' AND p.creator_id = ANY($2::bigint[]))
  )
ORDER BY p.created_at DESC
OFFSET $3 LIMIT $4
`

type GetPersonalizedFeedParams struct {
	UserID  int64
	Column2 []int64
	Offset  int32
	Limit   int32
}

type GetPersonalizedFeedRow struct {
	ID              int64
	PostBody        string
	CreatorID       int64
	CommentsCount   int32
	ReactionsCount  int32
	LastCommentedAt pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	LikedByUser     bool
	Image           int64
}

func (q *Queries) GetPersonalizedFeed(ctx context.Context, arg GetPersonalizedFeedParams) ([]GetPersonalizedFeedRow, error) {
	rows, err := q.db.Query(ctx, getPersonalizedFeed,
		arg.UserID,
		arg.Column2,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetPersonalizedFeedRow{}
	for rows.Next() {
		var i GetPersonalizedFeedRow
		if err := rows.Scan(
			&i.ID,
			&i.PostBody,
			&i.CreatorID,
			&i.CommentsCount,
			&i.ReactionsCount,
			&i.LastCommentedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LikedByUser,
			&i.Image,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPublicFeed = `-- name: GetPublicFeed :many
SELECT
    p.id,
    p.post_body,
    p.creator_id,
    p.comments_count,
    p.reactions_count,
    p.last_commented_at,
    p.created_at,
    p.updated_at,

    EXISTS (
        SELECT 1 FROM reactions r
        WHERE r.content_id = p.id
          AND r.user_id = $1
          AND r.deleted_at IS NULL
    ) AS liked_by_user,

COALESCE(
    (SELECT i.id
     FROM images i
     WHERE i.parent_id = p.id AND i.deleted_at IS NULL
     ORDER BY i.sort_order ASC
     LIMIT 1
    ), 0
)::bigint AS image

   
FROM posts p


WHERE p.deleted_at IS NULL
  AND p.audience = 'everyone'
ORDER BY p.created_at DESC
OFFSET $2 LIMIT $3
`

type GetPublicFeedParams struct {
	UserID int64
	Offset int32
	Limit  int32
}

type GetPublicFeedRow struct {
	ID              int64
	PostBody        string
	CreatorID       int64
	CommentsCount   int32
	ReactionsCount  int32
	LastCommentedAt pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	LikedByUser     bool
	Image           int64
}

func (q *Queries) GetPublicFeed(ctx context.Context, arg GetPublicFeedParams) ([]GetPublicFeedRow, error) {
	rows, err := q.db.Query(ctx, getPublicFeed, arg.UserID, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetPublicFeedRow{}
	for rows.Next() {
		var i GetPublicFeedRow
		if err := rows.Scan(
			&i.ID,
			&i.PostBody,
			&i.CreatorID,
			&i.CommentsCount,
			&i.ReactionsCount,
			&i.LastCommentedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LikedByUser,
			&i.Image,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserPostsPaginated = `-- name: GetUserPostsPaginated :many

SELECT
    p.id,
    p.post_body,
    p.creator_id,
    p.comments_count,
    p.reactions_count,
    p.last_commented_at,
    p.created_at,
    p.updated_at,

    EXISTS (    -- Has the requesting user liked the post?
        SELECT 1 FROM reactions r
        WHERE r.content_id = p.id
          AND r.user_id = $2
          AND r.deleted_at IS NULL
    ) AS liked_by_user,

COALESCE(
    (SELECT i.id
     FROM images i
     WHERE i.parent_id = p.id AND i.deleted_at IS NULL
     ORDER BY i.sort_order ASC
     LIMIT 1
    ), 0
)::bigint AS image   

  
FROM posts p



WHERE p.creator_id = $1                      -- target user we are viewing
  AND p.group_id IS NULL                     -- exclude group posts
  AND p.deleted_at IS NULL

  AND (                    
        p.creator_id = $2    -- If viewer *is* the creator — show all posts                
        OR p.audience = 'everyone' -- followers must be checked in users service
        OR (
            p.audience = 'selected'            -- must be specifically allowed
            AND EXISTS (
                SELECT 1
                FROM post_audience pa
                WHERE pa.post_id = p.id
                  AND pa.allowed_user_id = $2
            )
        )
         OR (p.audience = 'followers' AND $3::bool = TRUE)
     )

GROUP BY p.id
ORDER BY p.created_at DESC
LIMIT $4 OFFSET $5
`

type GetUserPostsPaginatedParams struct {
	CreatorID int64
	UserID    int64
	Column3   bool
	Limit     int32
	Offset    int32
}

type GetUserPostsPaginatedRow struct {
	ID              int64
	PostBody        string
	CreatorID       int64
	CommentsCount   int32
	ReactionsCount  int32
	LastCommentedAt pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	LikedByUser     bool
	Image           int64
}

// pagination
func (q *Queries) GetUserPostsPaginated(ctx context.Context, arg GetUserPostsPaginatedParams) ([]GetUserPostsPaginatedRow, error) {
	rows, err := q.db.Query(ctx, getUserPostsPaginated,
		arg.CreatorID,
		arg.UserID,
		arg.Column3,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetUserPostsPaginatedRow{}
	for rows.Next() {
		var i GetUserPostsPaginatedRow
		if err := rows.Scan(
			&i.ID,
			&i.PostBody,
			&i.CreatorID,
			&i.CommentsCount,
			&i.ReactionsCount,
			&i.LastCommentedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.LikedByUser,
			&i.Image,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
