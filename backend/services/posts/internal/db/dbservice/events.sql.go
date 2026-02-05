package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createEvent = `-- name: CreateEvent :one

INSERT INTO events (
    event_title,
    event_body,
    event_creator_id,
    group_id,
    event_date
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id
`

type CreateEventParams struct {
	EventTitle     string
	EventBody      string
	EventCreatorID int64
	GroupID        int64
	EventDate      pgtype.Date
}

// inserts a new event and returns the id
func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) (int64, error) {
	row := q.db.QueryRow(ctx, createEvent,
		arg.EventTitle,
		arg.EventBody,
		arg.EventCreatorID,
		arg.GroupID,
		arg.EventDate,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const deleteEvent = `-- name: DeleteEvent :execrows
UPDATE events
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND event_creator_id=$2 AND deleted_at IS NULL
`

type DeleteEventParams struct {
	ID             int64
	EventCreatorID int64
}

// deletes an event with given id and creator id as long as it's not already deleted
//
// returns rows affected
//
// no rows could mean no event was found fitting the given criteria or was already marked deleted
func (q *Queries) DeleteEvent(ctx context.Context, arg DeleteEventParams) (int64, error) {
	result, err := q.db.Exec(ctx, deleteEvent, arg.ID, arg.EventCreatorID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const deleteEventResponse = `-- name: DeleteEventResponse :execrows
UPDATE event_responses
SET deleted_at = CURRENT_TIMESTAMP
WHERE event_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type DeleteEventResponseParams struct {
	EventID int64
	UserID  int64
}

// soft-deletes an event response to an event with given id, by user with given id, as long as it's not already marked deleted
//
// returns rows affected
//
// 0 rows could mean no response fitting the criteria was found, or was already marked deleted
func (q *Queries) DeleteEventResponse(ctx context.Context, arg DeleteEventResponseParams) (int64, error) {
	result, err := q.db.Exec(ctx, deleteEventResponse, arg.EventID, arg.UserID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const editEvent = `-- name: EditEvent :execrows
UPDATE events
SET event_title = $1,
    event_body = $2,
    event_date = $3
WHERE id = $4 AND event_creator_id=$5 AND deleted_at IS NULL
`

type EditEventParams struct {
	EventTitle     string
	EventBody      string
	EventDate      pgtype.Date
	ID             int64
	EventCreatorID int64
}

// updates event title, body and date for event with given id and given creator id, as long as it wasn't marked deleted
//
// returns rows affected
//
// 0 rows affected could mean no event response fitting the criteria was found, or it was marked deleted
func (q *Queries) EditEvent(ctx context.Context, arg EditEventParams) (int64, error) {
	result, err := q.db.Exec(ctx, editEvent,
		arg.EventTitle,
		arg.EventBody,
		arg.EventDate,
		arg.ID,
		arg.EventCreatorID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getEventsByGroupId = `-- name: GetEventsByGroupId :many
SELECT
    e.id,
    e.event_title,
    e.event_body,
    e.event_creator_id,
    e.group_id,
    e.event_date,
    e.created_at,
    e.updated_at,
    e.going_count,
    e.not_going_count,

    COALESCE(
        (
            SELECT i.id
            FROM images i
            WHERE i.parent_id = e.id
              AND i.deleted_at IS NULL
            ORDER BY i.sort_order ASC
            LIMIT 1
        ),
        0
    )::bigint AS image,

    -- user response (NULL if no response)
    er.going AS user_response

FROM events e
LEFT JOIN event_responses er
    ON er.event_id = e.id
   AND er.user_id = $4
   AND er.deleted_at IS NULL

WHERE e.group_id = $1
  AND e.deleted_at IS NULL
  AND e.event_date >= CURRENT_DATE

ORDER BY e.event_date DESC
OFFSET $2
LIMIT $3
`

type GetEventsByGroupIdParams struct {
	GroupID int64
	Offset  int32
	Limit   int32
	UserID  int64
}

type GetEventsByGroupIdRow struct {
	ID             int64
	EventTitle     string
	EventBody      string
	EventCreatorID int64
	GroupID        int64
	EventDate      pgtype.Date
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	GoingCount     int32
	NotGoingCount  int32
	Image          int64
	UserResponse   pgtype.Bool
}

// returns paginated events for given group id, ordered by descending event date
//
// includes going/not going count, and requester's response if any
func (q *Queries) GetEventsByGroupId(ctx context.Context, arg GetEventsByGroupIdParams) ([]GetEventsByGroupIdRow, error) {
	rows, err := q.db.Query(ctx, getEventsByGroupId,
		arg.GroupID,
		arg.Offset,
		arg.Limit,
		arg.UserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetEventsByGroupIdRow{}
	for rows.Next() {
		var i GetEventsByGroupIdRow
		if err := rows.Scan(
			&i.ID,
			&i.EventTitle,
			&i.EventBody,
			&i.EventCreatorID,
			&i.GroupID,
			&i.EventDate,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.GoingCount,
			&i.NotGoingCount,
			&i.Image,
			&i.UserResponse,
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

const upsertEventResponse = `-- name: UpsertEventResponse :execrows
INSERT INTO event_responses (event_id, user_id, going)
VALUES ($1, $2, $3)
ON CONFLICT (event_id, user_id)
DO UPDATE
SET
    going = EXCLUDED.going,
    deleted_at = NULL,
    updated_at = CURRENT_TIMESTAMP
`

type UpsertEventResponseParams struct {
	EventID int64
	UserID  int64
	Going   bool
}

// inserts a new event response for given event id and user id, or updates it if it already exists
func (q *Queries) UpsertEventResponse(ctx context.Context, arg UpsertEventResponseParams) (int64, error) {
	result, err := q.db.Exec(ctx, upsertEventResponse, arg.EventID, arg.UserID, arg.Going)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
