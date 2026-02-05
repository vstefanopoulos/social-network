package events

import (
	"context"
	"social-network/services/notifications/internal/application"
	pb "social-network/shared/gen-go/notifications"
)

// EventProcessor defines the interface for processing notification events
type EventProcessor interface {
	Handle(ctx context.Context, event *pb.NotificationEvent) error
}

// ApplicationService defines the interface for the application layer
type ApplicationService interface {
	CreatePostCommentNotification(ctx context.Context, userID, commenterID, postID int64, commenterUsername, commentContent string, aggregate bool) error
	CreatePostLikeNotification(ctx context.Context, userID, likerID, postID int64, likerUsername string, aggregate bool) error
	CreateFollowRequestNotification(ctx context.Context, targetUserID, requesterUserID int64, requesterUsername string) error
	CreateNewFollowerNotification(ctx context.Context, targetUserID, followerUserID int64, followerUsername string, aggregate bool) error
	CreateGroupInviteNotification(ctx context.Context, invitedUserID, inviterUserID, groupID int64, groupName, inviterUsername string) error
	CreateGroupInviteForMultipleUsers(ctx context.Context, invitedUserIDs []int64, inviterUserID, groupID int64, groupName, inviterUsername string) error
	CreateGroupJoinRequestNotification(ctx context.Context, groupOwnerID, requesterID, groupID int64, groupName, requesterUsername string) error
	CreateNewEventNotification(ctx context.Context, userID, eventCreatorID, groupID, eventID int64, groupName, eventTitle string) error
	CreateNewEventForMultipleUsers(ctx context.Context, userIDs []int64, eventCreatorID int64, groupID, eventID int64, groupName, eventTitle string) error
	CreateMentionNotification(ctx context.Context, userID, mentionerID, postID int64, mentionerUsername, postContent, mentionText string) error
	CreateNewMessageNotification(ctx context.Context, userID, senderID, chatID int64, senderUsername, messageContent string, aggregate bool) error
	CreateNewMessageForMultipleUsers(ctx context.Context, userIDs []int64, senderID, chatID int64, senderUsername, messageContent string, aggregate bool) error
	CreateFollowRequestAcceptedNotification(ctx context.Context, requesterUserID, targetUserID int64, targetUsername string) error
	CreateFollowRequestRejectedNotification(ctx context.Context, requesterUserID, targetUserID int64, targetUsername string) error
	CreateGroupInviteAcceptedNotification(ctx context.Context, inviterUserID, invitedUserID, groupID int64, invitedUsername, groupName string) error
	CreateGroupInviteRejectedNotification(ctx context.Context, inviterUserID, invitedUserID, groupID int64, invitedUsername, groupName string) error
	CreateGroupJoinRequestAcceptedNotification(ctx context.Context, requesterUserID, groupOwnerID, groupID int64, groupName string) error
	CreateGroupJoinRequestRejectedNotification(ctx context.Context, requesterUserID, groupOwnerID, groupID int64, groupName string) error
	DeleteFollowRequestNotification(ctx context.Context, targetUserID, requesterUserID int64) error
	DeleteGroupJoinRequestNotification(ctx context.Context, groupOwnerID, requesterUserID, groupID int64) error
	CreateDefaultNotificationTypes(ctx context.Context) error
	GetNotification(ctx context.Context, notificationID, userID int64) (*application.Notification, error)
	GetUserNotifications(ctx context.Context, userID int64, limit, offset int32) ([]*application.Notification, error)
	GetUserNotificationsCount(ctx context.Context, userID int64) (int64, error)
	GetUserUnreadNotificationsCount(ctx context.Context, userID int64) (int64, error)
	MarkNotificationAsRead(ctx context.Context, notificationID, userID int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	DeleteNotification(ctx context.Context, notificationID, userID int64) error
	CreateNotification(ctx context.Context, userID int64, notifType application.NotificationType, title, message, sourceService string, sourceEntityID int64, needsAction bool, payload map[string]string) (*application.Notification, error)
	CreateNotificationWithAggregation(ctx context.Context, userID int64, notifType application.NotificationType, title, message, sourceService string, sourceEntityID int64, needsAction bool, payload map[string]string, aggregate bool) (*application.Notification, error)
	CreateNotifications(ctx context.Context, notifications []struct {
		UserID         int64
		Type           application.NotificationType
		Title          string
		Message        string
		SourceService  string
		SourceEntityID int64
		NeedsAction    bool
		Payload        map[string]string
	}) ([]*application.Notification, error)
}
