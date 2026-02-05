package dbservice

import (
	"context"
)

const suggestUsersByPostActivity = `-- name: SuggestUsersByPostActivity :many
WITH

u1 AS (
    SELECT 
        r.user_id,
        4 AS score
    FROM reactions r
    JOIN posts p ON p.id = r.content_id
    WHERE p.creator_id = $1
      AND p.audience = 'everyone'
      AND r.user_id <> $1
),

u2 AS (
    SELECT
        c.comment_creator_id AS user_id,
        4 AS score
    FROM comments c
    JOIN posts p ON p.id = c.parent_id
    WHERE p.creator_id = $1
      AND p.audience = 'everyone'
      AND c.comment_creator_id <> $1
),

u3 AS (
    SELECT DISTINCT
        r2.user_id,
        3 AS score
    FROM reactions r1                        -- your likes
    JOIN reactions r2 ON r1.content_id = r2.content_id
    WHERE r1.user_id = $1
      AND r2.user_id <> $1
),

u4 AS (
    SELECT DISTINCT
        c2.comment_creator_id AS user_id,
        2 AS score
    FROM comments c1                         -- your comments
    JOIN comments c2 ON c1.parent_id = c2.parent_id
    WHERE c1.comment_creator_id = $1
      AND c2.comment_creator_id <> $1
),

combined AS (
    SELECT user_id, SUM(score) AS total_score
    FROM (
        SELECT user_id, score FROM u1
        UNION ALL
        SELECT user_id, score FROM u2
        UNION ALL
        SELECT user_id, score FROM u3
        UNION ALL
        SELECT user_id, score FROM u4
    ) scored
    GROUP BY user_id
)

SELECT user_id
FROM combined
ORDER BY total_score DESC, random()
LIMIT 5
`

// U1: Users who liked one or more of *your public posts*
// U2: Users who commented on your public posts
// U3: Users who liked the same posts as you
// U4: Users who commented on the same posts as you
// Combine scores
func (q *Queries) SuggestUsersByPostActivity(ctx context.Context, creatorID int64) ([]int64, error) {
	rows, err := q.db.Query(ctx, suggestUsersByPostActivity, creatorID)
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
