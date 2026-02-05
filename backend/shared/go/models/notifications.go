package models

import (
	"social-network/shared/gen-go/notifications"
	ct "social-network/shared/go/ct"
)

type CreateNotificationRequest struct {
	UserId         ct.Id
	Title          string
	Message        string
	Type           notifications.NotificationType
	SourceService  string
	SourceEntityId int64
	NeedsAction    bool
	Payload        map[string]string
	Aggregate      bool
}
