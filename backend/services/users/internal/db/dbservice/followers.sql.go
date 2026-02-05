package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const acceptFollowRequest = `-- name: AcceptFollowRequest :exec
WITH updated AS (
    UPDATE follow_requests
    SET status = 'accepted', updated_at = NOW()
    WHERE requester_id = $1
      AND target_id    = $2
      AND deleted_at IS NULL
    RETURNING requester_id, target_id
)
INSERT INTO follows (follower_id, following_id, created_at,deleted_at)
SELECT requester_id, target_id, NOW(), NULL
FROM updated
ON CONFLICT  (follower_id, following_id)
DO UPDATE SET deleted_at=NULL;
`

type AcceptFollowRequestParams struct {
	RequesterID int64
	TargetID    int64
}

func (q *Queries) AcceptFollowRequest(ctx context.Context, arg AcceptFollowRequestParams) error {
	_, err := q.db.Exec(ctx, acceptFollowRequest, arg.RequesterID, arg.TargetID)
	return err
}

const areFollowingEachOther = `-- name: AreFollowingEachOther :one
WITH u1 AS (
  SELECT EXISTS (
    SELECT 1
    FROM follows f
    WHERE f.follower_id = $1 
	AND f.following_id = $2
	AND f.deleted_at IS NULL
  ) AS user1_follows_user2
),
u2 AS (
  SELECT EXISTS (
    SELECT 1
    FROM follows f
    WHERE f.follower_id = $2 
	AND f.following_id = $1
	AND f.deleted_at IS NULL
  ) AS user2_follows_user1
)
SELECT
  u1.user1_follows_user2,
  u2.user2_follows_user1
FROM u1, u2
`

type AreFollowingEachOtherParams struct {
	FollowerID  int64
	FollowingID int64
}

type AreFollowingEachOtherRow struct {
	User1FollowsUser2 bool
	User2FollowsUser1 bool
}

func (q *Queries) AreFollowingEachOther(ctx context.Context, arg AreFollowingEachOtherParams) (AreFollowingEachOtherRow, error) {
	row := q.db.QueryRow(ctx, areFollowingEachOther, arg.FollowerID, arg.FollowingID)
	var i AreFollowingEachOtherRow
	err := row.Scan(&i.User1FollowsUser2, &i.User2FollowsUser1)
	return i, err
}

const followUser = `-- name: FollowUser :one
SELECT follow_user($1, $2)
`

type FollowUserParams struct {
	PFollower int64
	PTarget   int64
}

func (q *Queries) FollowUser(ctx context.Context, arg FollowUserParams) (string, error) {
	row := q.db.QueryRow(ctx, followUser, arg.PFollower, arg.PTarget)
	var follow_user string
	err := row.Scan(&follow_user)
	return follow_user, err
}

const getFollowSuggestions = `-- name: GetFollowSuggestions :many
WITH
s1 AS (
    SELECT 
        f2.following_id AS user_id,
        5 AS score         -- weighted higher
    FROM follows f1 
    JOIN follows f2 ON f1.following_id = f2.follower_id
    WHERE f1.follower_id = $1
	  AND f1.deleted_at IS NULL
	  AND f2.deleted_at IS NULL
      AND f2.following_id <> $1
      AND NOT EXISTS (
          SELECT 1 FROM follows x
          WHERE x.follower_id = $1 
		  AND x.following_id = f2.following_id
		  AND x.deleted_at IS NULL
      )
),

s2 AS (
    SELECT
        gm2.user_id,
        3 AS score        -- lighter weight than follows
    FROM group_members gm1
    JOIN group_members gm2 ON gm1.group_id = gm2.group_id
    WHERE gm1.user_id = $1
      AND gm2.user_id <> $1
      AND gm2.deleted_at IS NULL
),

combined AS (
    SELECT user_id, SUM(score) AS total_score
    FROM (
        SELECT user_id, score FROM s1
        UNION ALL
        SELECT user_id, score FROM s2
    ) scored
    GROUP BY user_id
)

SELECT 
    u.id,
    u.username,
    u.avatar_id,
    c.total_score
FROM combined c
JOIN users u ON u.id = c.user_id
ORDER BY c.total_score DESC, random()
LIMIT 5
`

type GetFollowSuggestionsRow struct {
	ID         int64
	Username   string
	AvatarID   int64
	TotalScore int64
}

// S1: second-degree follows
// S2: shared groups
// Combine & score
func (q *Queries) GetFollowSuggestions(ctx context.Context, followerID int64) ([]GetFollowSuggestionsRow, error) {
	rows, err := q.db.Query(ctx, getFollowSuggestions, followerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetFollowSuggestionsRow{}
	for rows.Next() {
		var i GetFollowSuggestionsRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.AvatarID,
			&i.TotalScore,
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

const getFollowerCount = `-- name: GetFollowerCount :one
SELECT COUNT(*) 
FROM follows
WHERE following_id = $1
AND deleted_at IS NULL;
`

func (q *Queries) GetFollowerCount(ctx context.Context, followingID int64) (int64, error) {
	row := q.db.QueryRow(ctx, getFollowerCount, followingID)
	var follower_count int64
	err := row.Scan(&follower_count)
	return follower_count, err
}

const getFollowers = `-- name: GetFollowers :many
SELECT u.id, u.username, u.avatar_id,u.profile_public, f.created_at AS followed_at
FROM follows f
JOIN users u ON u.id = f.follower_id
WHERE f.following_id = $1
AND f.deleted_at IS NULL
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3
`

type GetFollowersParams struct {
	FollowingID int64
	Limit       int32
	Offset      int32
}

type GetFollowersRow struct {
	ID            int64
	Username      string
	AvatarID      int64
	ProfilePublic bool
	FollowedAt    pgtype.Timestamptz
}

func (q *Queries) GetFollowers(ctx context.Context, arg GetFollowersParams) ([]GetFollowersRow, error) {
	rows, err := q.db.Query(ctx, getFollowers, arg.FollowingID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetFollowersRow{}
	for rows.Next() {
		var i GetFollowersRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.AvatarID,
			&i.ProfilePublic,
			&i.FollowedAt,
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

const getFollowing = `-- name: GetFollowing :many
SELECT u.id, u.username,u.avatar_id,u.profile_public, f.created_at AS followed_at
FROM follows f
JOIN users u ON u.id = f.following_id
WHERE f.follower_id = $1
AND f.deleted_at IS NULL
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3
`

type GetFollowingParams struct {
	FollowerID int64
	Limit      int32
	Offset     int32
}

type GetFollowingRow struct {
	ID            int64
	Username      string
	AvatarID      int64
	ProfilePublic bool
	FollowedAt    pgtype.Timestamptz
}

func (q *Queries) GetFollowing(ctx context.Context, arg GetFollowingParams) ([]GetFollowingRow, error) {
	rows, err := q.db.Query(ctx, getFollowing, arg.FollowerID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetFollowingRow{}
	for rows.Next() {
		var i GetFollowingRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.AvatarID,
			&i.ProfilePublic,
			&i.FollowedAt,
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

const getFollowingCount = `-- name: GetFollowingCount :one
SELECT COUNT(*) 
FROM follows
WHERE follower_id = $1
AND deleted_at IS NULL;
`

func (q *Queries) GetFollowingCount(ctx context.Context, followerID int64) (int64, error) {
	row := q.db.QueryRow(ctx, getFollowingCount, followerID)
	var following_count int64
	err := row.Scan(&following_count)
	return following_count, err
}

const getFollowingIds = `-- name: GetFollowingIds :many
SELECT following_id
FROM follows 
WHERE follower_id = $1
AND deleted_at IS NULL;
`

func (q *Queries) GetFollowingIds(ctx context.Context, followerID int64) ([]int64, error) {
	rows, err := q.db.Query(ctx, getFollowingIds, followerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []int64{}
	for rows.Next() {
		var following_id int64
		if err := rows.Scan(&following_id); err != nil {
			return nil, err
		}
		items = append(items, following_id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getMutualFollowers = `-- name: GetMutualFollowers :many
SELECT u.id, u.username
FROM follows f1
JOIN follows f2 ON f1.follower_id = f2.follower_id
JOIN users u ON u.id = f1.follower_id
WHERE f1.following_id = $1
  AND f2.following_id = $2
  AND f1.deleted_at IS NULL
  AND f2.deleted_at IS NULL;
`

type GetMutualFollowersParams struct {
	FollowingID   int64
	FollowingID_2 int64
}

type GetMutualFollowersRow struct {
	ID       int64
	Username string
}

func (q *Queries) GetMutualFollowers(ctx context.Context, arg GetMutualFollowersParams) ([]GetMutualFollowersRow, error) {
	rows, err := q.db.Query(ctx, getMutualFollowers, arg.FollowingID, arg.FollowingID_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetMutualFollowersRow{}
	for rows.Next() {
		var i GetMutualFollowersRow
		if err := rows.Scan(&i.ID, &i.Username); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const isFollowRequestPending = `-- name: IsFollowRequestPending :one
SELECT EXISTS(
    SELECT 1 
    FROM follow_requests
    WHERE requester_id = $1
      AND target_id = $2
      AND status = 'pending'
) AS has_pending_request
`

type IsFollowRequestPendingParams struct {
	RequesterID int64
	TargetID    int64
}

func (q *Queries) IsFollowRequestPending(ctx context.Context, arg IsFollowRequestPendingParams) (bool, error) {
	row := q.db.QueryRow(ctx, isFollowRequestPending, arg.RequesterID, arg.TargetID)
	var has_pending_request bool
	err := row.Scan(&has_pending_request)
	return has_pending_request, err
}

const isFollowing = `-- name: IsFollowing :one
SELECT EXISTS (
    SELECT 1 FROM follows
    WHERE follower_id =$1 
	AND following_id = $2
	AND deleted_at IS NULL
);
`

type IsFollowingParams struct {
	FollowerID  int64
	FollowingID int64
}

func (q *Queries) IsFollowing(ctx context.Context, arg IsFollowingParams) (bool, error) {
	row := q.db.QueryRow(ctx, isFollowing, arg.FollowerID, arg.FollowingID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const rejectFollowRequest = `-- name: RejectFollowRequest :exec
UPDATE follow_requests
SET status = 'rejected', updated_at = NOW()
WHERE requester_id = $1 AND target_id = $2
`

type RejectFollowRequestParams struct {
	RequesterID int64
	TargetID    int64
}

func (q *Queries) RejectFollowRequest(ctx context.Context, arg RejectFollowRequestParams) error {
	_, err := q.db.Exec(ctx, rejectFollowRequest, arg.RequesterID, arg.TargetID)
	return err
}

const unfollowUser = `-- name: UnfollowUser :one
WITH updated_follow AS (
  UPDATE follows
  SET deleted_at = NOW()
  WHERE follower_id = $1
    AND following_id = $2
    AND deleted_at IS NULL
  RETURNING 1
),
deleted_request AS (
  DELETE FROM follow_requests
  WHERE requester_id = $1
    AND target_id = $2
  RETURNING 1
)
SELECT
  CASE
    WHEN EXISTS (SELECT 1 FROM updated_follow) THEN 'unfollow'
    WHEN EXISTS (SELECT 1 FROM deleted_request) THEN 'cancel_request'
    ELSE 'none'
  END AS action;
`

type UnfollowUserParams struct {
	FollowerID  int64
	FollowingID int64
}

func (q *Queries) UnfollowUser(ctx context.Context, arg UnfollowUserParams) (string, error) {
	var action string
	err := q.db.QueryRow(ctx, unfollowUser, arg.FollowerID, arg.FollowingID).
		Scan(&action)
	return action, err
}
