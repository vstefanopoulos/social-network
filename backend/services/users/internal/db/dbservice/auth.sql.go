package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const banUser = `-- name: BanUser :exec
UPDATE users
SET 
    current_status = 'banned',
    ban_ends_at = $2
WHERE id = $1
`

type BanUserParams struct {
	ID        int64
	BanEndsAt pgtype.Timestamptz
}

func (q *Queries) BanUser(ctx context.Context, arg BanUserParams) error {
	_, err := q.db.Exec(ctx, banUser, arg.ID, arg.BanEndsAt)
	return err
}

const getUserForLogin = `-- name: GetUserForLogin :one
SELECT
    u.id,
    u.username,
    u.avatar_id,
    u.profile_public,
    au.password_hash
FROM users u
JOIN auth_user au ON au.user_id = u.id
WHERE (u.username = $1 OR au.email = $1) 
  AND password_hash = $2
  AND u.current_status = 'active'
  AND u.deleted_at IS NULL
`

type GetUserForLoginParams struct {
	Username     string
	PasswordHash string
}

type GetUserForLoginRow struct {
	ID            int64
	Username      string
	AvatarID      int64
	ProfilePublic bool
	PasswordHash  string
}

func (q *Queries) GetUserForLogin(ctx context.Context, arg GetUserForLoginParams) (GetUserForLoginRow, error) {
	row := q.db.QueryRow(ctx, getUserForLogin, arg.Username, arg.PasswordHash)
	var i GetUserForLoginRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.AvatarID,
		&i.ProfilePublic,
		&i.PasswordHash,
	)
	return i, err
}

const getUserPassword = `-- name: GetUserPassword :one
SELECT 
    password_hash
FROM
    auth_user
WHERE user_id=$1
`

func (q *Queries) GetUserPassword(ctx context.Context, userID int64) (string, error) {
	row := q.db.QueryRow(ctx, getUserPassword, userID)
	var password_hash string
	err := row.Scan(&password_hash)
	return password_hash, err
}

const insertNewUser = `-- name: InsertNewUser :one
INSERT INTO users (
    username,
    first_name,
    last_name,
    date_of_birth,
    avatar_id,
    about_me,
    profile_public
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id
`

type InsertNewUserParams struct {
	Username      string
	FirstName     string
	LastName      string
	DateOfBirth   pgtype.Date
	AvatarID      int64
	AboutMe       string
	ProfilePublic bool
}

func (q *Queries) InsertNewUser(ctx context.Context, arg InsertNewUserParams) (int64, error) {
	row := q.db.QueryRow(ctx, insertNewUser,
		arg.Username,
		arg.FirstName,
		arg.LastName,
		arg.DateOfBirth,
		arg.AvatarID,
		arg.AboutMe,
		arg.ProfilePublic,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const insertNewUserAuth = `-- name: InsertNewUserAuth :exec
INSERT INTO auth_user (
    user_id,
    email,
    password_hash
) VALUES (
       $1, $2, $3
)
`

type InsertNewUserAuthParams struct {
	UserID       int64
	Email        string
	PasswordHash string
}

func (q *Queries) InsertNewUserAuth(ctx context.Context, arg InsertNewUserAuthParams) error {
	_, err := q.db.Exec(ctx, insertNewUserAuth, arg.UserID, arg.Email, arg.PasswordHash)
	return err
}

const softDeleteUser = `-- name: SoftDeleteUser :exec
UPDATE users
SET
    current_status = 'deleted',
    deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND deleted_at IS NULL
`

func (q *Queries) SoftDeleteUser(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, softDeleteUser, id)
	return err
}

const unbanUser = `-- name: UnbanUser :exec
UPDATE users
SET 
    current_status = 'active',
    ban_ends_at = NULL
WHERE id = $1
`

func (q *Queries) UnbanUser(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, unbanUser, id)
	return err
}
