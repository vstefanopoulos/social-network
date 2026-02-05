package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createComment = `-- name: CreateComment :one
INSERT INTO comments (comment_creator_id, parent_id, comment_body)
VALUES ($1, $2, $3)
RETURNING id
`

type CreateCommentParams struct {
	CommentCreatorID int64
	ParentID         int64
	CommentBody      string
}

// inserts a new comment and returns the id
func (q *Queries) CreateComment(ctx context.Context, arg CreateCommentParams) (int64, error) {
	row := q.db.QueryRow(ctx, createComment, arg.CommentCreatorID, arg.ParentID, arg.CommentBody)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const deleteComment = `-- name: DeleteComment :execrows
UPDATE comments
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND comment_creator_id=$2 AND deleted_at IS NULL
`

type DeleteCommentParams struct {
	ID               int64
	CommentCreatorID int64
}

// soft-deletes a comment with given id and creator id, as long as it's not already marked deleted
// returns rows affected
// 0 rows could mean no comment fitting these criteria was found, or it was already deleted
func (q *Queries) DeleteComment(ctx context.Context, arg DeleteCommentParams) (int64, error) {
	result, err := q.db.Exec(ctx, deleteComment, arg.ID, arg.CommentCreatorID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const editComment = `-- name: EditComment :execrows
UPDATE comments
SET comment_body = $1
WHERE id = $2 AND comment_creator_id=$3 AND deleted_at IS NULL
`

type EditCommentParams struct {
	CommentBody      string
	ID               int64
	CommentCreatorID int64
}

// updates the body of a comment with given id and creator id, as long as it's not marked deleted
// returns rows affected
// 0 rows could mean no comment fitting the criteria was found, or it was already marked deleted
func (q *Queries) EditComment(ctx context.Context, arg EditCommentParams) (int64, error) {
	result, err := q.db.Exec(ctx, editComment, arg.CommentBody, arg.ID, arg.CommentCreatorID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getCommentsByPostId = `-- name: GetCommentsByPostId :many
SELECT
    c.id,
    c.comment_creator_id,
    c.comment_body,
    c.reactions_count,
    c.created_at,
    c.updated_at,

    EXISTS (
        SELECT 1
        FROM reactions r
        WHERE r.content_id = c.id
          AND r.user_id = $2
          AND r.deleted_at IS NULL
    ) AS liked_by_user,


    COALESCE(
    (SELECT i.id
     FROM images i
     WHERE i.parent_id = c.id AND i.deleted_at IS NULL
     ORDER BY i.sort_order ASC
     LIMIT 1
    ), 0
    
)::bigint AS image

FROM comments c
WHERE c.parent_id = $1
  AND c.deleted_at IS NULL
ORDER BY c.created_at DESC 
OFFSET $3
LIMIT $4
`

type GetCommentsByPostIdParams struct {
	ParentID int64
	UserID   int64
	Offset   int32
	Limit    int32
}

type GetCommentsByPostIdRow struct {
	ID               int64
	CommentCreatorID int64
	CommentBody      string
	ReactionsCount   int32
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	LikedByUser      bool
	Image            int64
}

// returns paginated comments of post with given id, in descending created order
func (q *Queries) GetCommentsByPostId(ctx context.Context, arg GetCommentsByPostIdParams) ([]GetCommentsByPostIdRow, error) {
	rows, err := q.db.Query(ctx, getCommentsByPostId,
		arg.ParentID,
		arg.UserID,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetCommentsByPostIdRow{}
	for rows.Next() {
		var i GetCommentsByPostIdRow
		if err := rows.Scan(
			&i.ID,
			&i.CommentCreatorID,
			&i.CommentBody,
			&i.ReactionsCount,
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

const getLatestCommentforPostId = `-- name: GetLatestCommentforPostId :one
SELECT
    c.id,
    c.comment_creator_id,
    c.parent_id,
    c.comment_body,
    c.reactions_count,
    c.created_at,
    c.updated_at,

    EXISTS (
        SELECT 1 FROM reactions r
        WHERE r.content_id = c.id
          AND r.user_id = $2
          AND r.deleted_at IS NULL
    ) AS liked_by_user,


    COALESCE(
    (SELECT i.id
     FROM images i
     WHERE i.parent_id = c.id AND i.deleted_at IS NULL
     ORDER BY i.sort_order ASC
     LIMIT 1
    ), 0

)::bigint AS image


FROM comments c
WHERE c.parent_id = $1
  AND c.deleted_at IS NULL
ORDER BY c.created_at DESC
LIMIT 1
`

type GetLatestCommentforPostIdParams struct {
	ParentID int64
	UserID   int64
}

type GetLatestCommentforPostIdRow struct {
	ID               int64
	CommentCreatorID int64
	ParentID         int64
	CommentBody      string
	ReactionsCount   int32
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	LikedByUser      bool
	Image            int64
}

func (q *Queries) GetLatestCommentforPostId(ctx context.Context, arg GetLatestCommentforPostIdParams) (GetLatestCommentforPostIdRow, error) {
	row := q.db.QueryRow(ctx, getLatestCommentforPostId, arg.ParentID, arg.UserID)
	var i GetLatestCommentforPostIdRow
	err := row.Scan(
		&i.ID,
		&i.CommentCreatorID,
		&i.ParentID,
		&i.CommentBody,
		&i.ReactionsCount,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.LikedByUser,
		&i.Image,
	)
	return i, err
}

const getPostAudienceForComment = `
SELECT audience
FROM posts
WHERE id = $1;
`

func (q *Queries) GetPostAudienceForComment(ctx context.Context, postID int64) (string, error) {
	row := q.db.QueryRow(ctx, getPostAudienceForComment, postID)
	var audience string
	err := row.Scan(&audience)
	if err != nil {
		return "", err
	}
	return audience, nil
}
