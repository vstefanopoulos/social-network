package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const clearPostAudience = `-- name: ClearPostAudience :exec
DELETE FROM post_audience
WHERE post_id = $1
`

func (q *Queries) ClearPostAudience(ctx context.Context, postID int64) error {
	_, err := q.db.Exec(ctx, clearPostAudience, postID)
	return err
}

const createPost = `-- name: CreatePost :one
INSERT INTO posts (post_body, creator_id, group_id, audience)
VALUES ($1, $2, $3, $4)
RETURNING id
`

type CreatePostParams struct {
	PostBody  string
	CreatorID int64
	GroupID   pgtype.Int8
	Audience  IntendedAudience
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (int64, error) {
	row := q.db.QueryRow(ctx, createPost,
		arg.PostBody,
		arg.CreatorID,
		arg.GroupID,
		arg.Audience,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const deletePost = `-- name: DeletePost :execrows
UPDATE posts
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND creator_id=$2 AND deleted_at IS NULL
`

type DeletePostParams struct {
	ID        int64
	CreatorID int64
}

func (q *Queries) DeletePost(ctx context.Context, arg DeletePostParams) (int64, error) {
	result, err := q.db.Exec(ctx, deletePost, arg.ID, arg.CreatorID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const editPostContent = `-- name: EditPostContent :execrows
UPDATE posts
SET post_body  = $1
WHERE id = $2 AND creator_id = $3 AND deleted_at IS NULL
`

type EditPostContentParams struct {
	PostBody  string
	ID        int64
	CreatorID int64
}

func (q *Queries) EditPostContent(ctx context.Context, arg EditPostContentParams) (int64, error) {
	result, err := q.db.Exec(ctx, editPostContent, arg.PostBody, arg.ID, arg.CreatorID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getMostPopularPostInGroup = `-- name: GetMostPopularPostInGroup :one
SELECT
    p.id,
    p.post_body,
    p.creator_id,
    COALESCE(p.group_id, 0)::bigint AS group_id,
    p.audience,
    p.comments_count,
    p.reactions_count,
    p.last_commented_at,
    p.created_at,
    p.updated_at,

    COALESCE(
        (
            SELECT i.id
            FROM images i
            WHERE i.parent_id = p.id
              AND i.deleted_at IS NULL
            ORDER BY i.sort_order ASC
            LIMIT 1
        ),
        0
    )::bigint AS image,


    (p.reactions_count + p.comments_count) AS popularity_score     -- popularity metric (likes + comments)

FROM posts p
WHERE p.group_id = $1
  AND p.deleted_at IS NULL

ORDER BY popularity_score DESC, p.created_at DESC
LIMIT 1
`

type GetMostPopularPostInGroupRow struct {
	ID              int64
	PostBody        string
	CreatorID       int64
	GroupID         int64
	Audience        IntendedAudience
	CommentsCount   int32
	ReactionsCount  int32
	LastCommentedAt pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	Image           int64
	PopularityScore int32
}

func (q *Queries) GetMostPopularPostInGroup(ctx context.Context, groupID pgtype.Int8) (GetMostPopularPostInGroupRow, error) {
	row := q.db.QueryRow(ctx, getMostPopularPostInGroup, groupID)
	var i GetMostPopularPostInGroupRow
	err := row.Scan(
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
		&i.Image,
		&i.PopularityScore,
	)
	return i, err
}

const getPostAudience = `-- name: GetPostAudience :many
SELECT allowed_user_id
FROM post_audience
WHERE post_id = $1
ORDER BY allowed_user_id
`

func (q *Queries) GetPostAudience(ctx context.Context, postID int64) ([]int64, error) {
	rows, err := q.db.Query(ctx, getPostAudience, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []int64{}
	for rows.Next() {
		var allowed_user_id int64
		if err := rows.Scan(&allowed_user_id); err != nil {
			return nil, err
		}
		items = append(items, allowed_user_id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPostByID = `-- name: GetPostByID :one
SELECT
    p.id,
    p.post_body,
    p.creator_id,
    COALESCE(p.group_id, 0)::bigint AS group_id,
    p.audience,
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
)::bigint AS image,

 COALESCE(
        (
            SELECT array_agg(pa.allowed_user_id ORDER BY pa.allowed_user_id)
            FROM post_audience pa
            WHERE pa.post_id = p.id
              AND p.audience = 'selected'
        ),
        ARRAY[]::bigint[]
    ) AS selected_audience


FROM posts p
WHERE p.id=$2
  AND p.deleted_at IS NULL
`

type GetPostByIDParams struct {
	UserID int64
	ID     int64
}

type GetPostByIDRow struct {
	ID               int64
	PostBody         string
	CreatorID        int64
	GroupID          int64
	Audience         IntendedAudience
	CommentsCount    int32
	ReactionsCount   int32
	LastCommentedAt  pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	LikedByUser      bool
	Image            int64
	SelectedAudience []int64
}

func (q *Queries) GetPostByID(ctx context.Context, arg GetPostByIDParams) (GetPostByIDRow, error) {
	row := q.db.QueryRow(ctx, getPostByID, arg.UserID, arg.ID)
	var i GetPostByIDRow
	err := row.Scan(
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
		&i.SelectedAudience,
	)
	return i, err
}

const insertPostAudience = `-- name: InsertPostAudience :execrows
INSERT INTO post_audience (post_id, allowed_user_id)
SELECT $1::bigint,
       allowed_user_id
FROM unnest($2::bigint[]) AS allowed_user_id
`

type InsertPostAudienceParams struct {
	PostID         int64
	AllowedUserIds []int64
}

func (q *Queries) InsertPostAudience(ctx context.Context, arg InsertPostAudienceParams) (int64, error) {
	result, err := q.db.Exec(ctx, insertPostAudience, arg.PostID, arg.AllowedUserIds)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const updatePostAudience = `-- name: UpdatePostAudience :execrows
UPDATE posts
SET audience = $3,
    updated_at = NOW()
WHERE 
    id = $1
    AND creator_id = $2
    AND deleted_at IS NULL
    AND (audience IS DISTINCT FROM $3)
`

type UpdatePostAudienceParams struct {
	ID        int64
	CreatorID int64
	Audience  IntendedAudience
}

func (q *Queries) UpdatePostAudience(ctx context.Context, arg UpdatePostAudienceParams) (int64, error) {
	result, err := q.db.Exec(ctx, updatePostAudience, arg.ID, arg.CreatorID, arg.Audience)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getBasicPostByID = `-- name: GetBasicPostByID :one
SELECT
    id,
    post_body,
    creator_id,
    COALESCE(group_id, 0)::bigint AS group_id,
    audience
      
FROM posts
WHERE id=$1
  AND deleted_at IS NULL
`

type GetBasicPostByIDRow struct {
	ID        int64
	PostBody  string
	CreatorID int64
	GroupID   int64
	Audience  IntendedAudience
}

func (q *Queries) GetBasicPostByID(ctx context.Context, postId int64) (GetBasicPostByIDRow, error) {
	row := q.db.QueryRow(ctx, getBasicPostByID, postId)
	var i GetBasicPostByIDRow
	err := row.Scan(
		&i.ID,
		&i.PostBody,
		&i.CreatorID,
		&i.GroupID,
		&i.Audience,
	)
	return i, err
}
