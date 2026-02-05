package mapping

import (
	pb "social-network/shared/gen-go/notifications"
	ct "social-network/shared/go/ct"
	"time"
)

type Notification struct {
	ID             ct.Id             `json:"id"`
	UserID         ct.Id             `json:"user_id"`
	Type           string            `json:"type"`
	SourceService  string            `json:"source_service"`
	SourceEntityID ct.Id             `json:"source_entity_id"`
	Seen           bool              `json:"seen"`
	NeedsAction    bool              `json:"needs_action"`
	Acted          bool              `json:"acted"`
	Count          int32             `json:"count"`
	Payload        map[string]string `json:"payload"`
	CreatedAt      time.Time         `json:"created_at"`
	ExpiresAt      time.Time         `json:"expires_at,omitzero"`
	DeletedAt      time.Time         `json:"deleted_at,omitzero"`
	Title          string            `json:"title"`
	Message        string            `json:"message"`
}

func PbToNotification(n *pb.Notification) Notification {

	notification := Notification{
		ID:             ct.Id(n.Id),
		UserID:         ct.Id(n.UserId),
		Type:           n.Type,
		SourceService:  n.SourceService,
		SourceEntityID: ct.Id(n.SourceEntityId),
		NeedsAction:    n.NeedsAction,
		Acted:          n.Acted,
		Count:          n.Count,
		Payload:        n.Payload,
		CreatedAt:      n.CreatedAt.AsTime(),
		ExpiresAt:      n.ExpiresAt.AsTime(),
		Title:          n.Title,
		Message:        n.Message,
	}

	if n.Status == pb.NotificationStatus_NOTIFICATION_STATUS_READ {
		notification.Seen = true
	}
	return notification
}

func PbToNotifications(ns []*pb.Notification) []Notification {
	var notifs []Notification
	for _, n := range ns {
		notifs = append(notifs, PbToNotification(n))
	}
	return notifs
}
