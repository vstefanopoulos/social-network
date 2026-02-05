package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getBatchUsersBasic = `-- name: GetBatchUsersBasic :many
SELECT
  id,
  username,
  avatar_id
FROM users
WHERE id = ANY($1::bigint[])
`

type GetBatchUsersBasicRow struct {
	ID       int64
	Username string
	AvatarID int64
}

func (q *Queries) GetBatchUsersBasic(ctx context.Context, dollar_1 []int64) ([]GetBatchUsersBasicRow, error) {
	rows, err := q.db.Query(ctx, getBatchUsersBasic, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetBatchUsersBasicRow{}
	for rows.Next() {
		var i GetBatchUsersBasicRow
		if err := rows.Scan(&i.ID, &i.Username, &i.AvatarID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserBasic = `-- name: GetUserBasic :one
SELECT
    id,
    username,
    avatar_id
FROM users
WHERE id = $1
`

type GetUserBasicRow struct {
	ID       int64
	Username string
	AvatarID int64
}

func (q *Queries) GetUserBasic(ctx context.Context, id int64) (GetUserBasicRow, error) {
	row := q.db.QueryRow(ctx, getUserBasic, id)
	var i GetUserBasicRow
	err := row.Scan(&i.ID, &i.Username, &i.AvatarID)
	return i, err
}

const getUserProfile = `-- name: GetUserProfile :one
SELECT
    u.id,
    u.username,
    u.first_name,
    u.last_name,
    u.date_of_birth,
    u.avatar_id,
    u.about_me,
    u.profile_public,
    u.created_at,
    a.email
FROM users u
INNER JOIN auth_user a
    ON a.user_id = u.id
WHERE u.id = $1
  AND u.deleted_at IS NULL
`

type GetUserProfileRow struct {
	ID            int64
	Username      string
	FirstName     string
	LastName      string
	DateOfBirth   pgtype.Date
	AvatarID      int64
	AboutMe       string
	ProfilePublic bool
	CreatedAt     pgtype.Timestamptz
	Email         string
}

func (q *Queries) GetUserProfile(ctx context.Context, id int64) (GetUserProfileRow, error) {
	row := q.db.QueryRow(ctx, getUserProfile, id)
	var i GetUserProfileRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.DateOfBirth,
		&i.AvatarID,
		&i.AboutMe,
		&i.ProfilePublic,
		&i.CreatedAt,
		&i.Email,
	)
	return i, err
}

const searchUsers = `-- name: SearchUsers :many
SELECT
    id,
    username,
    avatar_id,
    profile_public
FROM users
WHERE deleted_at IS NULL
  AND (
        CASE
            -- 3+ characters: fuzzy search (pg_trgm)
            WHEN LENGTH($1) >= 3 THEN
                username::text % $1
                OR first_name % $1
                OR last_name % $1

            -- 1–2 characters: prefix match WITHOUT LIKE
            ELSE
                LEFT(username::text, LENGTH($1)) = LOWER($1)
                OR LEFT(first_name, LENGTH($1)) = LOWER($1)
                OR LEFT(last_name, LENGTH($1)) = LOWER($1)
        END
      )
ORDER BY
    CASE
        -- 3+ characters: rank by similarity
        WHEN LENGTH($1) >= 3 THEN
            GREATEST(
                similarity(username::text, $1),
                similarity(first_name, $1),
                similarity(last_name, $1)
            )

        -- 1–2 characters: deterministic prefix ranking
        ELSE
            CASE
                WHEN LEFT(username::text, LENGTH($1)) = LOWER($1) THEN 3
                WHEN LEFT(first_name, LENGTH($1)) = LOWER($1) THEN 2
                WHEN LEFT(last_name, LENGTH($1)) = LOWER($1) THEN 1
                ELSE 0
            END
    END DESC,
    username ASC
LIMIT $2;
`

type SearchUsersParams struct {
	Query string
	Limit int32
}

type SearchUsersRow struct {
	ID            int64
	Username      string
	AvatarID      int64
	ProfilePublic bool
}

func (q *Queries) SearchUsers(ctx context.Context, arg SearchUsersParams) ([]SearchUsersRow, error) {
	rows, err := q.db.Query(ctx, searchUsers, arg.Query, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []SearchUsersRow{}
	for rows.Next() {
		var i SearchUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.AvatarID,
			&i.ProfilePublic,
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

const updateProfilePrivacy = `-- name: UpdateProfilePrivacy :exec
UPDATE users
SET profile_public=$2
WHERE id=$1
`

type UpdateProfilePrivacyParams struct {
	ID            int64
	ProfilePublic bool
}

func (q *Queries) UpdateProfilePrivacy(ctx context.Context, arg UpdateProfilePrivacyParams) error {
	_, err := q.db.Exec(ctx, updateProfilePrivacy, arg.ID, arg.ProfilePublic)
	return err
}

const updateUserEmail = `-- name: UpdateUserEmail :exec
UPDATE auth_user
SET
    email = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
`

type UpdateUserEmailParams struct {
	UserID int64
	Email  string
}

func (q *Queries) UpdateUserEmail(ctx context.Context, arg UpdateUserEmailParams) error {
	_, err := q.db.Exec(ctx, updateUserEmail, arg.UserID, arg.Email)
	return err
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE auth_user
SET
    password_hash = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1
`

type UpdateUserPasswordParams struct {
	UserID       int64
	PasswordHash string
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.Exec(ctx, updateUserPassword, arg.UserID, arg.PasswordHash)
	return err
}

const updateUserProfile = `-- name: UpdateUserProfile :one
UPDATE users
SET
    username      = $2,
    first_name    = $3,
    last_name     = $4,
    date_of_birth = $5,
    avatar_id        = $6,
    about_me      = $7,
    updated_at    = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, username, first_name, last_name, date_of_birth, avatar_id, about_me, profile_public, current_status, ban_ends_at, created_at, updated_at, deleted_at
`

type UpdateUserProfileParams struct {
	ID          int64
	Username    string
	FirstName   string
	LastName    string
	DateOfBirth pgtype.Date
	AvatarID    int64
	AboutMe     string
}

func (q *Queries) UpdateUserProfile(ctx context.Context, arg UpdateUserProfileParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserProfile,
		arg.ID,
		arg.Username,
		arg.FirstName,
		arg.LastName,
		arg.DateOfBirth,
		arg.AvatarID,
		arg.AboutMe,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.DateOfBirth,
		&i.AvatarID,
		&i.AboutMe,
		&i.ProfilePublic,
		&i.CurrentStatus,
		&i.BanEndsAt,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const removeImages = `-- name: RemoveImages :exec
UPDATE users
SET avatar_id = 0
WHERE avatar_id = ANY($1::bigint[]);
`

func (q *Queries) RemoveImages(ctx context.Context, arg []int64) error {
	_, err := q.db.Exec(ctx, removeImages, arg)
	return err
}
