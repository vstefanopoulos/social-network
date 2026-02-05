package dbservice

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type ContentType string

const (
	ContentTypePost    ContentType = "post"
	ContentTypeComment ContentType = "comment"
	ContentTypeEvent   ContentType = "event"
)

func (e *ContentType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ContentType(s)
	case string:
		*e = ContentType(s)
	default:
		return fmt.Errorf("unsupported scan type for ContentType: %T", src)
	}
	return nil
}

type NullContentType struct {
	ContentType ContentType
	Valid       bool // Valid is true if ContentType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullContentType) Scan(value interface{}) error {
	if value == nil {
		ns.ContentType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ContentType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullContentType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ContentType), nil
}

func (e ContentType) Valid() bool {
	switch e {
	case ContentTypePost,
		ContentTypeComment,
		ContentTypeEvent:
		return true
	}
	return false
}

type IntendedAudience string

const (
	IntendedAudienceEveryone  IntendedAudience = "everyone"
	IntendedAudienceFollowers IntendedAudience = "followers"
	IntendedAudienceSelected  IntendedAudience = "selected"
	IntendedAudienceGroup     IntendedAudience = "group"
)

func (e *IntendedAudience) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = IntendedAudience(s)
	case string:
		*e = IntendedAudience(s)
	default:
		return fmt.Errorf("unsupported scan type for IntendedAudience: %T", src)
	}
	return nil
}

type NullIntendedAudience struct {
	IntendedAudience IntendedAudience
	Valid            bool // Valid is true if IntendedAudience is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullIntendedAudience) Scan(value interface{}) error {
	if value == nil {
		ns.IntendedAudience, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.IntendedAudience.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullIntendedAudience) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.IntendedAudience), nil
}

func (e IntendedAudience) Valid() bool {
	switch e {
	case IntendedAudienceEveryone,
		IntendedAudienceFollowers,
		IntendedAudienceSelected,
		IntendedAudienceGroup:
		return true
	}
	return false
}

type Comment struct {
	ID               int64
	CommentCreatorID int64
	ParentID         int64
	CommentBody      string
	ReactionsCount   int32
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

type Event struct {
	ID             int64
	EventTitle     string
	EventBody      string
	EventCreatorID int64
	GroupID        int64
	EventDate      pgtype.Date
	GoingCount     int32
	NotGoingCount  int32
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

type EventResponse struct {
	ID        int64
	EventID   int64
	UserID    int64
	Going     bool
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

type Image struct {
	ID        int64
	ParentID  int64
	SortOrder int32
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

type MasterIndex struct {
	ID          int64
	ContentType ContentType
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

type Post struct {
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
	DeletedAt       pgtype.Timestamptz
}

type PostAudience struct {
	PostID        int64
	AllowedUserID int64
}

type Reaction struct {
	ID        int64
	ContentID int64
	UserID    int64
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}
