package dbservice

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type FollowRequestStatus string

const (
	FollowRequestStatusPending  FollowRequestStatus = "pending"
	FollowRequestStatusAccepted FollowRequestStatus = "accepted"
	FollowRequestStatusRejected FollowRequestStatus = "rejected"
)

func (e *FollowRequestStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = FollowRequestStatus(s)
	case string:
		*e = FollowRequestStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for FollowRequestStatus: %T", src)
	}
	return nil
}

type NullFollowRequestStatus struct {
	FollowRequestStatus FollowRequestStatus
	Valid               bool // Valid is true if FollowRequestStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullFollowRequestStatus) Scan(value interface{}) error {
	if value == nil {
		ns.FollowRequestStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.FollowRequestStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullFollowRequestStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.FollowRequestStatus), nil
}

func (e FollowRequestStatus) Valid() bool {
	switch e {
	case FollowRequestStatusPending,
		FollowRequestStatusAccepted,
		FollowRequestStatusRejected:
		return true
	}
	return false
}

type GroupInviteStatus string

const (
	GroupInviteStatusPending  GroupInviteStatus = "pending"
	GroupInviteStatusAccepted GroupInviteStatus = "accepted"
	GroupInviteStatusDeclined GroupInviteStatus = "declined"
	GroupInviteStatusExpired  GroupInviteStatus = "expired"
)

func (e *GroupInviteStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = GroupInviteStatus(s)
	case string:
		*e = GroupInviteStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for GroupInviteStatus: %T", src)
	}
	return nil
}

type NullGroupInviteStatus struct {
	GroupInviteStatus GroupInviteStatus
	Valid             bool // Valid is true if GroupInviteStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullGroupInviteStatus) Scan(value interface{}) error {
	if value == nil {
		ns.GroupInviteStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.GroupInviteStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullGroupInviteStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.GroupInviteStatus), nil
}

func (e GroupInviteStatus) Valid() bool {
	switch e {
	case GroupInviteStatusPending,
		GroupInviteStatusAccepted,
		GroupInviteStatusDeclined,
		GroupInviteStatusExpired:
		return true
	}
	return false
}

type GroupRole string

const (
	GroupRoleMember GroupRole = "member"
	GroupRoleOwner  GroupRole = "owner"
)

func (e *GroupRole) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = GroupRole(s)
	case string:
		*e = GroupRole(s)
	default:
		return fmt.Errorf("unsupported scan type for GroupRole: %T", src)
	}
	return nil
}

type NullGroupRole struct {
	GroupRole GroupRole
	Valid     bool // Valid is true if GroupRole is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullGroupRole) Scan(value interface{}) error {
	if value == nil {
		ns.GroupRole, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.GroupRole.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullGroupRole) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.GroupRole), nil
}

func (e GroupRole) Valid() bool {
	switch e {
	case GroupRoleMember,
		GroupRoleOwner:
		return true
	}
	return false
}

type JoinRequestStatus string

const (
	JoinRequestStatusPending  JoinRequestStatus = "pending"
	JoinRequestStatusAccepted JoinRequestStatus = "accepted"
	JoinRequestStatusRejected JoinRequestStatus = "rejected"
)

func (e *JoinRequestStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = JoinRequestStatus(s)
	case string:
		*e = JoinRequestStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for JoinRequestStatus: %T", src)
	}
	return nil
}

type NullJoinRequestStatus struct {
	JoinRequestStatus JoinRequestStatus
	Valid             bool // Valid is true if JoinRequestStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullJoinRequestStatus) Scan(value interface{}) error {
	if value == nil {
		ns.JoinRequestStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.JoinRequestStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullJoinRequestStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.JoinRequestStatus), nil
}

func (e JoinRequestStatus) Valid() bool {
	switch e {
	case JoinRequestStatusPending,
		JoinRequestStatusAccepted,
		JoinRequestStatusRejected:
		return true
	}
	return false
}

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBanned  UserStatus = "banned"
	UserStatusDeleted UserStatus = "deleted"
)

func (e *UserStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UserStatus(s)
	case string:
		*e = UserStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for UserStatus: %T", src)
	}
	return nil
}

type NullUserStatus struct {
	UserStatus UserStatus
	Valid      bool // Valid is true if UserStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullUserStatus) Scan(value interface{}) error {
	if value == nil {
		ns.UserStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.UserStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullUserStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.UserStatus), nil
}

func (e UserStatus) Valid() bool {
	switch e {
	case UserStatusActive,
		UserStatusBanned,
		UserStatusDeleted:
		return true
	}
	return false
}

type AuthUser struct {
	UserID       int64
	Email        string
	PasswordHash string
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	LastLoginAt  pgtype.Timestamptz
}

type Follow struct {
	FollowerID  int64
	FollowingID int64
	CreatedAt   pgtype.Timestamptz
}

type FollowRequest struct {
	RequesterID int64
	TargetID    int64
	Status      FollowRequestStatus
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

type Group struct {
	ID               int64
	GroupOwner       int64
	GroupTitle       string
	GroupDescription string
	GroupImageID     int64
	MembersCount     int32
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

type GroupInvite struct {
	GroupID    int64
	SenderID   int64
	ReceiverID int64
	Status     GroupInviteStatus
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

type GroupJoinRequest struct {
	GroupID   int64
	UserID    int64
	Status    JoinRequestStatus
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

type GroupMember struct {
	GroupID   int64
	UserID    int64
	Role      NullGroupRole
	JoinedAt  pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

type User struct {
	ID            int64
	Username      string
	FirstName     string
	LastName      string
	DateOfBirth   pgtype.Date
	AvatarID      int64
	AboutMe       string
	ProfilePublic bool
	CurrentStatus UserStatus
	BanEndsAt     pgtype.Timestamptz
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
}
