package application

import (
	ct "social-network/shared/go/ct"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	FollowRequest            NotificationType = "follow_request"
	NewFollower              NotificationType = "new_follower"
	GroupInvite              NotificationType = "group_invite"
	GroupJoinRequest         NotificationType = "group_join_request"
	NewEvent                 NotificationType = "new_event"
	PostLike                 NotificationType = "like"
	PostComment              NotificationType = "post_reply"
	Mention                  NotificationType = "mention"
	NewMessage               NotificationType = "new_message"
	FollowRequestAccepted    NotificationType = "follow_request_accepted"
	FollowRequestRejected    NotificationType = "follow_request_rejected"
	GroupInviteAccepted      NotificationType = "group_invite_accepted"
	GroupInviteRejected      NotificationType = "group_invite_rejected"
	GroupJoinRequestAccepted NotificationType = "group_join_request_accepted"
	GroupJoinRequestRejected NotificationType = "group_join_request_rejected"
)

// Notification represents a notification entity
type Notification struct {
	ID             ct.Id             `json:"id"`
	UserID         ct.Id             `json:"user_id"`
	Type           NotificationType  `json:"type"`
	SourceService  string            `json:"source_service"`
	SourceEntityID ct.Id             `json:"source_entity_id"`
	Seen           bool              `json:"seen"`
	NeedsAction    bool              `json:"needs_action"`
	Acted          bool              `json:"acted"`
	Count          int32             `json:"count"`
	Payload        map[string]string `json:"payload"`
	CreatedAt      time.Time         `json:"created_at"`
	ExpiresAt      *time.Time        `json:"expires_at"`
	DeletedAt      *time.Time        `json:"deleted_at"`
	Title          string            `json:"title"`
	Message        string            `json:"message"`
}
