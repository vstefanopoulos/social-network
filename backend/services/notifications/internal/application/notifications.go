package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	db "social-network/services/notifications/internal/db/sqlc"
	pb "social-network/shared/gen-go/notifications"
	ct "social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateNotification creates a new notification
func (a *Application) CreateNotification(ctx context.Context, userID int64, notifType NotificationType, title, message, sourceService string, sourceEntityID int64, needsAction bool, payload map[string]string) (*Notification, error) {
	return a.CreateNotificationWithAggregation(ctx, userID, notifType, title, message, sourceService, sourceEntityID, needsAction, payload, false)
}

// CreateNotificationWithAggregation creates a new notification or aggregates with an existing one if applicable
func (a *Application) CreateNotificationWithAggregation(ctx context.Context, userID int64, notifType NotificationType, title, message, sourceService string, sourceEntityID int64, needsAction bool, payload map[string]string, aggregate bool) (*Notification, error) {
	if !aggregate {
		// If aggregation is disabled, create a new notification as before
		return a.createNotification(ctx, userID, notifType, title, message, sourceService, sourceEntityID, needsAction, payload, 1)
	}

	// If aggregation is enabled, first check for an existing unread notification of same type and entity
	existingNotification, err := a.DB.GetUnreadNotificationByTypeAndEntity(ctx, db.GetUnreadNotificationByTypeAndEntityParams{
		UserID:         userID,
		NotifType:      string(notifType),
		SourceEntityID: pgtype.Int8{Int64: sourceEntityID, Valid: true},
	})

	if err != nil {
		// If no existing notification found (which is normal), create a new one
		if errors.Is(err, pgx.ErrNoRows) {
			return a.createNotification(ctx, userID, notifType, title, message, sourceService, sourceEntityID, needsAction, payload, 1)
		}
		return nil, fmt.Errorf("failed to check for existing notification: %w", err)
	}

	// If an existing unread notification is found, increment its count and update it
	newCount := existingNotification.Count.Int32 + 1
	err = a.DB.UpdateNotificationCount(ctx, db.UpdateNotificationCountParams{
		Count:  pgtype.Int4{Int32: newCount, Valid: true},
		ID:     existingNotification.ID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update notification count: %w", err)
	}

	// Fetch and return the updated notification
	updatedNotification, err := a.DB.GetNotificationByID(ctx, existingNotification.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated notification: %w", err)
	}

	// Convert database notification to our model
	notification := &Notification{
		ID:            ct.Id(updatedNotification.ID),
		UserID:        ct.Id(updatedNotification.UserID),
		Type:          NotificationType(updatedNotification.NotifType),
		SourceService: updatedNotification.SourceService,
		Title:         a.formatAggregatedTitle(title, int64(newCount)),
		Message:       a.formatAggregatedMessage(message, int64(newCount)),
		Count:         newCount,
	}

	// Handle optional fields with proper type conversion
	if updatedNotification.SourceEntityID.Valid {
		notification.SourceEntityID = ct.Id(updatedNotification.SourceEntityID.Int64)
	}
	notification.Seen = updatedNotification.Seen.Bool
	notification.NeedsAction = updatedNotification.NeedsAction.Bool
	notification.Acted = updatedNotification.Acted.Bool

	if updatedNotification.CreatedAt.Valid {
		notification.CreatedAt = updatedNotification.CreatedAt.Time
	}
	if updatedNotification.ExpiresAt.Valid {
		notification.ExpiresAt = &updatedNotification.ExpiresAt.Time
	}
	if updatedNotification.DeletedAt.Valid {
		notification.DeletedAt = &updatedNotification.DeletedAt.Time
	}

	// Parse the payload JSON back to map
	if len(updatedNotification.Payload) > 0 {
		err = json.Unmarshal(updatedNotification.Payload, &notification.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	notification.Payload, err = encodePayloadIds(notification.Payload)
	if err != nil {
		tele.Error(ctx, "Error in encoding payload ids", "payload", notification.Payload)
		return nil, nil
	}

	// For response notifications, mark the related original notification as acted
	switch notification.Type {
	case FollowRequestAccepted, FollowRequestRejected, GroupInviteAccepted, GroupInviteRejected, GroupJoinRequestAccepted, GroupJoinRequestRejected:
		err = a.MarkRelatedNotificationAsActed(ctx, notification.Type, userID, sourceEntityID, payload)
		if err != nil {
			// Log the error but don't fail the notification creation
			tele.Error(ctx, "failed to mark related notification as acted: @1", "error", err.Error())
		}
	}

	// Publish the notification to NATS for real-time delivery to the live service
	// We do this asynchronously to not block the notification creation
	go func() {
		// Create a background context for the NATS publish operation
		natsCtx := context.Background()
		if err := a.publishNotificationToNATS(natsCtx, notification); err != nil {
			// Log the error but don't fail the notification creation
			tele.Error(natsCtx, "failed to publish notification to nats in background: @1", "error", err.Error())
		}
	}()

	return notification, nil
}

// createNotification is a helper function that creates a notification with a specific count
func (a *Application) createNotification(ctx context.Context, userID int64, notifType NotificationType, title, message, sourceService string, sourceEntityID int64, needsAction bool, payload map[string]string, count int32) (*Notification, error) {
	// Prepare the JSON payload
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Calculate expiration time (default 30 days if not specified)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	// Create the notification in the database
	dbNotification, err := a.DB.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:         userID,
		NotifType:      string(notifType),
		SourceService:  sourceService,
		SourceEntityID: pgtype.Int8{Int64: sourceEntityID, Valid: true},
		NeedsAction:    pgtype.Bool{Bool: needsAction, Valid: true},
		Acted:          pgtype.Bool{Bool: false, Valid: true}, // New notifications haven't been acted upon yet
		Payload:        payloadJSON,
		ExpiresAt:      pgtype.Timestamptz{Time: expiresAt, Valid: true},
		Count:          pgtype.Int4{Int32: count, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Convert database notification to our model
	notification := &Notification{
		ID:            ct.Id(dbNotification.ID),
		UserID:        ct.Id(dbNotification.UserID),
		Type:          NotificationType(dbNotification.NotifType),
		SourceService: dbNotification.SourceService,
		Title:         title,
		Message:       message,
	}

	// Handle optional fields with proper type conversion
	if dbNotification.SourceEntityID.Valid {
		notification.SourceEntityID = ct.Id(dbNotification.SourceEntityID.Int64)
	}
	notification.Seen = dbNotification.Seen.Bool
	notification.NeedsAction = dbNotification.NeedsAction.Bool
	notification.Acted = dbNotification.Acted.Bool
	notification.Count = dbNotification.Count.Int32 // Set the count from the database response

	if dbNotification.CreatedAt.Valid {
		notification.CreatedAt = dbNotification.CreatedAt.Time
	}
	if dbNotification.ExpiresAt.Valid {
		notification.ExpiresAt = &dbNotification.ExpiresAt.Time
	}
	if dbNotification.DeletedAt.Valid {
		notification.DeletedAt = &dbNotification.DeletedAt.Time
	}

	// Parse the payload JSON back to map
	if len(dbNotification.Payload) > 0 {
		err = json.Unmarshal(dbNotification.Payload, &notification.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	notification.Payload, err = encodePayloadIds(notification.Payload)
	if err != nil {
		tele.Error(ctx, "Error in encoding payload ids", "payload", notification.Payload)
		return nil, nil
	}

	// For response notifications, mark the related original notification as acted
	switch notification.Type {
	case FollowRequestAccepted, FollowRequestRejected, GroupInviteAccepted, GroupInviteRejected, GroupJoinRequestAccepted, GroupJoinRequestRejected:
		err = a.MarkRelatedNotificationAsActed(ctx, notification.Type, userID, sourceEntityID, payload)
		if err != nil {
			// Log the error but don't fail the notification creation
			tele.Error(ctx, "failed to mark related notification as acted: @1", "error", err.Error())
		}
	}

	// Publish the notification to NATS for real-time delivery to the live service
	// We do this asynchronously to not block the notification creation
	go func() {
		// Create a background context for the NATS publish operation
		natsCtx := context.Background()
		if err := a.publishNotificationToNATS(natsCtx, notification); err != nil {
			// Log the error but don't fail the notification creation
			tele.Error(natsCtx, "failed to publish notification to nats in background: @1", "error", err.Error())
		}
	}()

	return notification, nil
}

// formatAggregatedTitle formats notification titles when notifications are aggregated
func (a *Application) formatAggregatedTitle(originalTitle string, count int64) string {
	if count <= 1 {
		return originalTitle
	}

	// For now, we'll handle a few common cases, but this could be extended based on notification type
	switch originalTitle {
	case "Post Liked":
		return fmt.Sprintf("%d People Liked Your Post", count)
	case "New Comment":
		return fmt.Sprintf("%d People Commented On Your Post", count)
	case "New Follower":
		return fmt.Sprintf("%d New Followers", count)
	case "New Message":
		return fmt.Sprintf("%d New Messages", count)
	default:
		return fmt.Sprintf("%d Notifications", count)
	}
}

// formatAggregatedMessage formats notification messages when notifications are aggregated
func (a *Application) formatAggregatedMessage(originalMessage string, count int64) string {
	if count <= 1 {
		return originalMessage
	}

	// For now, we'll handle a few common cases, but this could be extended based on notification type
	switch {
	case strings.Contains(originalMessage, "liked your post"):
		return fmt.Sprintf("%d people liked your post", count)
	case strings.Contains(originalMessage, "commented on your post"):
		return fmt.Sprintf("%d people commented on your post", count)
	case strings.Contains(originalMessage, "is now following you"):
		return fmt.Sprintf("%d people are now following you", count)
	case strings.Contains(originalMessage, "sent you a message"):
		return fmt.Sprintf("%d people sent you a message", count)
	default:
		return fmt.Sprintf("You have %d notifications", count)
	}
}

// CreateNotifications creates multiple notifications in a batch
func (a *Application) CreateNotifications(ctx context.Context, notifications []struct {
	UserID         int64
	Type           NotificationType
	Title          string
	Message        string
	SourceService  string
	SourceEntityID int64
	NeedsAction    bool
	Payload        map[string]string
}) ([]*Notification, error) {
	result := make([]*Notification, 0, len(notifications))

	for _, n := range notifications {
		notification, err := a.CreateNotification(ctx, n.UserID, n.Type, n.Title, n.Message, n.SourceService, n.SourceEntityID, n.NeedsAction, n.Payload)
		if err != nil {
			return nil, err
		}
		result = append(result, notification)
	}

	return result, nil
}

// GetNotification retrieves a single notification by ID
func (a *Application) GetNotification(ctx context.Context, notificationID, userID int64) (*Notification, error) {
	dbNotification, err := a.DB.GetNotificationByID(ctx, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Only return the notification if it belongs to the user
	if dbNotification.UserID != userID {
		return nil, fmt.Errorf("notification not found")
	}

	notification := &Notification{
		ID:            ct.Id(dbNotification.ID),
		UserID:        ct.Id(dbNotification.UserID),
		Type:          NotificationType(dbNotification.NotifType),
		SourceService: dbNotification.SourceService,
	}

	// Handle optional fields with proper type conversion
	if dbNotification.SourceEntityID.Valid {
		notification.SourceEntityID = ct.Id(dbNotification.SourceEntityID.Int64)
	}
	notification.Seen = dbNotification.Seen.Bool
	notification.NeedsAction = dbNotification.NeedsAction.Bool
	notification.Acted = dbNotification.Acted.Bool

	if dbNotification.CreatedAt.Valid {
		notification.CreatedAt = dbNotification.CreatedAt.Time
	}
	if dbNotification.ExpiresAt.Valid {
		notification.ExpiresAt = &dbNotification.ExpiresAt.Time
	}
	if dbNotification.DeletedAt.Valid {
		notification.DeletedAt = &dbNotification.DeletedAt.Time
	}

	// Parse the payload JSON back to map
	if len(dbNotification.Payload) > 0 {
		err = json.Unmarshal(dbNotification.Payload, &notification.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	notification.Payload, err = encodePayloadIds(notification.Payload)
	if err != nil {
		tele.Error(ctx, "Error in encoding payload ids", "payload", notification.Payload)
		return nil, nil
	}

	return notification, nil
}

// GetUserNotifications retrieves notifications for a user
func (a *Application) GetUserNotifications(ctx context.Context, userID int64, limit, offset int32) ([]*Notification, error) {
	dbNotifications, err := a.DB.GetUserNotifications(ctx, db.GetUserNotificationsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user notifications: %w", err)
	}

	notifications := make([]*Notification, len(dbNotifications))
	for i, dbNotif := range dbNotifications {
		notification := &Notification{
			ID:            ct.Id(dbNotif.ID),
			UserID:        ct.Id(dbNotif.UserID),
			Type:          NotificationType(dbNotif.NotifType),
			SourceService: dbNotif.SourceService,
			Count:         dbNotif.Count.Int32,
		}

		// Handle optional fields with proper type conversion
		if dbNotif.SourceEntityID.Valid {
			notification.SourceEntityID = ct.Id(dbNotif.SourceEntityID.Int64)
		}
		notification.Seen = dbNotif.Seen.Bool
		notification.NeedsAction = dbNotif.NeedsAction.Bool
		notification.Acted = dbNotif.Acted.Bool

		if dbNotif.CreatedAt.Valid {
			notification.CreatedAt = dbNotif.CreatedAt.Time
		}
		if dbNotif.ExpiresAt.Valid {
			notification.ExpiresAt = &dbNotif.ExpiresAt.Time
		}
		if dbNotif.DeletedAt.Valid {
			notification.DeletedAt = &dbNotif.DeletedAt.Time
		}

		// Parse the payload JSON back to map
		if len(dbNotif.Payload) > 0 {
			err = json.Unmarshal(dbNotif.Payload, &notification.Payload)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
			}
		}

		notification.Payload, err = encodePayloadIds(notification.Payload)
		if err != nil {
			tele.Error(ctx, "Error in encoding payload ids", "payload", notification.Payload)
			return nil, nil
		}

		notifications[i] = notification
	}

	return notifications, nil
}

// GetUserNotificationsCount gets the total count of notifications for a user
func (a *Application) GetUserNotificationsCount(ctx context.Context, userID int64) (int64, error) {
	count, err := a.DB.GetUserNotificationsCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get notifications count: %w", err)
	}
	return count, nil
}

// GetUserUnreadNotificationsCount gets the count of unread notifications for a user
func (a *Application) GetUserUnreadNotificationsCount(ctx context.Context, userID int64) (int64, error) {
	count, err := a.DB.GetUserUnreadNotificationsCount(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread notifications count: %w", err)
	}
	return count, nil
}

// MarkNotificationAsRead marks a notification as read
func (a *Application) MarkNotificationAsRead(ctx context.Context, notificationID, userID int64) error {
	err := a.DB.MarkNotificationAsRead(ctx, db.MarkNotificationAsReadParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

// MarkNotificationAsActed marks a notification as acted
func (a *Application) MarkNotificationAsActed(ctx context.Context, notificationID, userID int64) error {
	err := a.DB.MarkNotificationAsActed(ctx, db.MarkNotificationAsActedParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark notification as acted: %w", err)
	}
	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (a *Application) MarkAllAsRead(ctx context.Context, userID int64) error {
	err := a.DB.MarkAllAsRead(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// DeleteNotification soft deletes a notification
func (a *Application) DeleteNotification(ctx context.Context, notificationID, userID int64) error {
	err := a.DB.DeleteNotification(ctx, db.DeleteNotificationParams{
		ID:     notificationID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	return nil
}

// CreateDefaultNotificationTypes ensures default notification types are in the database
func (a *Application) CreateDefaultNotificationTypes(ctx context.Context) error {
	defaultTypes := []struct {
		Type           string
		Category       string
		DefaultEnabled bool
	}{
		{string(FollowRequest), "social", true},
		{string(NewFollower), "social", true},
		{string(GroupInvite), "group", true},
		{string(GroupJoinRequest), "group", true},
		{string(NewEvent), "group", true},
		{string(PostLike), "posts", true},
		{string(PostComment), "posts", true},
		{string(Mention), "posts", true},
		{string(NewMessage), "chat", true},
		{string(FollowRequestAccepted), "social", true},
		{string(FollowRequestRejected), "social", true},
		{string(GroupInviteAccepted), "group", true},
		{string(GroupInviteRejected), "group", true},
		{string(GroupJoinRequestAccepted), "group", true},
		{string(GroupJoinRequestRejected), "group", true},
	}

	for _, nt := range defaultTypes {
		err := a.DB.CreateNotificationType(ctx, db.CreateNotificationTypeParams{
			NotifType:      nt.Type,
			Category:       pgtype.Text{String: nt.Category, Valid: true},
			DefaultEnabled: pgtype.Bool{Bool: nt.DefaultEnabled, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to create notification type %s: %w", nt.Type, err)
		}
	}

	return nil
}

// publishNotificationToNATS publishes a notification to NATS for real-time delivery to the live service
func (a *Application) publishNotificationToNATS(ctx context.Context, notification *Notification) error {
	if a.NatsConn == nil {
		tele.Warn(ctx, "NATS connection is nil, skipping notification publish")
		return nil
	}

	// Marshal the notification to JSON format (similar to chat service)
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification to JSON: %w", err)
	}

	// Publish to the user-specific NATS subject
	subject := ct.UserKey(notification.UserID)
	err = a.NatsConn.Publish(subject, notificationJSON)
	if err != nil {
		// Log the error but don't fail the notification creation
		tele.Error(ctx, "failed to publish notification to nats: @1", "error", err.Error())
		return fmt.Errorf("failed to publish notification to nats: %w", err)
	}

	// Flush to ensure the message is sent
	err = a.NatsConn.Flush()
	if err != nil {
		tele.Error(ctx, "failed to flush nats connection: @1", "error", err.Error())
	}

	tele.Info(ctx, "Published notification to nats for user @1", "userId", notification.UserID)
	return nil
}

// DeleteFollowRequestNotification deletes a follow request notification when the request is cancelled
func (a *Application) DeleteFollowRequestNotification(ctx context.Context, targetUserID, requesterUserID int64) error {
	// Find the specific follow request notification to delete
	// This looks for an existing notification of type FollowRequest where the source entity is the requester
	dbNotification, err := a.DB.GetUnreadNotificationByTypeAndEntity(ctx, db.GetUnreadNotificationByTypeAndEntityParams{
		UserID:         targetUserID,
		NotifType:      string(FollowRequest),
		SourceEntityID: pgtype.Int8{Int64: requesterUserID, Valid: true},
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If no notification exists, that's fine - nothing to delete
			tele.Info(ctx, "No follow request notification found to delete for user @1 from requester @2", "targetUserID", targetUserID, "requesterUserID", requesterUserID)
			return nil
		}
		return fmt.Errorf("failed to find follow request notification: %w", err)
	}

	notificationID := dbNotification.ID

	// Delete the notification
	err = a.DB.DeleteNotification(ctx, db.DeleteNotificationParams{
		ID:     notificationID,
		UserID: targetUserID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete follow request notification: %w", err)
	}

	// Publish notification deletion to NATS for real-time updates
	go func() {
		// Create a background context for the NATS publish operation
		natsCtx := context.Background()
		if err := a.publishNotificationDeletionToNATS(natsCtx, notificationID, targetUserID); err != nil {
			// Log the error but don't fail the notification deletion
			tele.Error(natsCtx, "failed to publish notification deletion to nats in background: @1", "error", err.Error())
		}
	}()

	tele.Info(ctx, "Deleted follow request notification for user @1 from requester @2", "targetUserID", targetUserID, "requesterUserID", requesterUserID)
	return nil
}

// DeleteGroupJoinRequestNotification deletes a group join request notification when the request is cancelled
func (a *Application) DeleteGroupJoinRequestNotification(ctx context.Context, groupOwnerID, requesterUserID, groupID int64) error {
	// Find the specific group join request notification to delete
	// This looks for an existing notification of type GroupJoinRequest where the source entity is the group
	dbNotification, err := a.DB.GetUnreadNotificationByTypeAndEntity(ctx, db.GetUnreadNotificationByTypeAndEntityParams{
		UserID:         groupOwnerID,
		NotifType:      string(GroupJoinRequest),
		SourceEntityID: pgtype.Int8{Int64: groupID, Valid: true},
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If no notification exists, that's fine - nothing to delete
			tele.Info(ctx, "No group join request notification found to delete for owner @1 from requester @2 for group @3", "groupOwnerID", groupOwnerID, "requesterUserID", requesterUserID, "groupID", groupID)
			return nil
		}
		return fmt.Errorf("failed to find group join request notification: %w", err)
	}

	// Check if the notification is for the correct requester by looking at the payload
	var payload map[string]string
	if len(dbNotification.Payload) > 0 {
		err = json.Unmarshal(dbNotification.Payload, &payload)
		if err != nil {
			return fmt.Errorf("failed to unmarshal notification payload: %w", err)
		}

		// Verify this notification is for the specific requester
		requesterIDStr, exists := payload["requester_id"]
		if !exists {
			return fmt.Errorf("requester_id not found in notification payload")
		}

		requesterID, err := strconv.ParseInt(requesterIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse requester_id from payload: %w", err)
		}

		if requesterID != requesterUserID {
			tele.Info(ctx, "Found group join request notification but requester doesn't match - ignoring")
			return nil
		}
	}

	notificationID := dbNotification.ID

	// Delete the notification
	err = a.DB.DeleteNotification(ctx, db.DeleteNotificationParams{
		ID:     notificationID,
		UserID: groupOwnerID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete group join request notification: %w", err)
	}

	// Publish notification deletion to NATS for real-time updates
	go func() {
		// Create a background context for the NATS publish operation
		natsCtx := context.Background()
		if err := a.publishNotificationDeletionToNATS(natsCtx, notificationID, groupOwnerID); err != nil {
			// Log the error but don't fail the notification deletion
			tele.Error(natsCtx, "failed to publish notification deletion to nats in background: @1", "error", err.Error())
		}
	}()

	tele.Info(ctx, "Deleted group join request notification for owner @1 from requester @2 for group @3", "groupOwnerID", groupOwnerID, "requesterUserID", requesterUserID, "groupID", groupID)
	return nil
}

// MarkRelatedNotificationAsActed marks the original request notification as acted when a response is created
func (a *Application) MarkRelatedNotificationAsActed(ctx context.Context, responseNotifType NotificationType, userID int64, sourceEntityID int64, payload map[string]string) error {
	// Determine the original notification type based on the response type
	var originalNotifType NotificationType
	var originalUserID int64
	var originalSourceEntityID int64

	switch responseNotifType {
	case FollowRequestAccepted, FollowRequestRejected:
		originalNotifType = FollowRequest
		// For follow requests, the original notification was sent to the target user (recipient of the follow request)
		// The source entity ID in the original notification is the requester's ID
		if requesterIDStr, exists := payload["requester_id"]; exists {
			requesterID, err := strconv.ParseInt(requesterIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse requester_id from payload: %w", err)
			}

			// The original notification was sent to the target user (from payload), with source entity as requester
			// The target user ID is the one who received the original follow request
			if targetIDStr, exists := payload["target_id"]; exists {
				targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
				if err != nil {
					return fmt.Errorf("failed to parse target_id from payload: %w", err)
				}

				originalUserID = targetID
				originalSourceEntityID = requesterID
			} else {
				return fmt.Errorf("target_id not found in payload for follow request response")
			}
		} else {
			return fmt.Errorf("requester_id not found in payload for follow request response")
		}
	case GroupInviteAccepted, GroupInviteRejected:
		originalNotifType = GroupInvite
		// For group invites, the original notification was sent to the invited user (recipient of the invite)
		// The source entity ID in the original notification is the group ID
		// The invited user ID is stored in the payload
		if invitedIDStr, exists := payload["invited_id"]; exists {
			invitedID, err := strconv.ParseInt(invitedIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse invited_id from payload: %w", err)
			}

			originalUserID = invitedID
			originalSourceEntityID = sourceEntityID
		} else {
			return fmt.Errorf("invited_id not found in payload for group invite response")
		}
	case GroupJoinRequestAccepted, GroupJoinRequestRejected:
		originalNotifType = GroupJoinRequest
		// For group join requests, the original notification was sent to the group owner (not the requester)
		// The source entity ID in the original notification is the group ID
		// The payload contains the group_owner_notification_user_id, which tells us who received the original notification
		if groupOwnerIDStr, exists := payload["group_owner_notification_user_id"]; exists {
			groupOwnerID, err := strconv.ParseInt(groupOwnerIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse group_owner_notification_user_id from payload: %w", err)
			}

			// The original notification was sent to the group owner (groupOwnerID)
			originalUserID = groupOwnerID
			originalSourceEntityID = sourceEntityID // group ID
		} else {
			return fmt.Errorf("group_owner_notification_user_id not found in payload for group join request response")
		}
	default:
		// This is not a response notification type that requires marking related notifications as acted
		return nil
	}

	// Find the original notification that should be marked as acted
	dbNotification, err := a.DB.GetNotificationByTypeAndEntity(ctx, db.GetNotificationByTypeAndEntityParams{
		UserID:         originalUserID,
		NotifType:      string(originalNotifType),
		SourceEntityID: pgtype.Int8{Int64: originalSourceEntityID, Valid: true},
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// If no notification exists, that's fine - nothing to mark as acted
			tele.Info(ctx, "No related notification found to mark as acted for user @1, type @2, source entity @3", "userID", originalUserID, "originalNotifType", originalNotifType, "originalSourceEntityID", originalSourceEntityID)
			return nil
		}
		return fmt.Errorf("failed to find related notification to mark as acted: %w", err)
	}

	// Mark the original notification as acted
	err = a.DB.MarkNotificationAsActed(ctx, db.MarkNotificationAsActedParams{
		ID:     dbNotification.ID,
		UserID: originalUserID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark related notification as acted: %w", err)
	}

	tele.Info(ctx, "Marked related notification as acted for user @1, type @2, source entity @3", "userID", originalUserID, "originalNotifType", originalNotifType, "originalSourceEntityID", originalSourceEntityID)
	return nil
}

// publishNotificationDeletionToNATS publishes a notification deletion to NATS for real-time delivery to the live service
func (a *Application) publishNotificationDeletionToNATS(ctx context.Context, notificationID, userID int64) error {
	if a.NatsConn == nil {
		tele.Warn(ctx, "NATS connection is nil, skipping notification deletion publish")
		return nil
	}

	// Create a notification deletion message
	deletionMsg := &pb.NotificationDeletion{
		NotificationId: notificationID,
		UserId:         userID,
		DeletedAt:      timestamppb.Now(),
	}

	// Marshal the deletion message to JSON format
	deletionJSON, err := json.Marshal(deletionMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal notification deletion to JSON: %w", err)
	}

	// Publish to the user-specific NATS subject for notification deletions
	subject := ct.UserKey(userID)
	err = a.NatsConn.Publish(subject, deletionJSON)
	if err != nil {
		// Log the error but don't fail the notification deletion
		tele.Error(ctx, "failed to publish notification deletion to nats: @1", "error", err.Error())
		return fmt.Errorf("failed to publish notification deletion to nats: %w", err)
	}

	// Flush to ensure the message is sent
	err = a.NatsConn.Flush()
	if err != nil {
		tele.Error(ctx, "failed to flush nats connection: @1", "error", err.Error())
	}

	tele.Info(ctx, "Published notification deletion to nats for user @1, notification @2", "userId", userID, "notificationId", notificationID)
	return nil
}

func encodePayloadIds(payload map[string]string) (map[string]string, error) {
	for key, value := range payload {
		if strings.HasSuffix(key, "_id") {
			id, err := strconv.Atoi(value)
			if err != nil {
				return nil, err

			}
			encryptedId, err := ct.EncodeId(ct.Id(id))
			if err != nil {
				return nil, err
			}
			payload[key] = encryptedId
		}
	}
	return payload, nil
}
