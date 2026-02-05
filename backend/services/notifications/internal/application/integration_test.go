package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"social-network/services/notifications/internal/db/sqlc"
)

// Test that all notification trigger functions work
func TestNotificationTriggerFunctions(t *testing.T) {
	mockDB := new(MockDB)
	app := NewApplicationWithMocks(mockDB)

	ctx := context.Background()

	// Test CreateGroupInviteNotification
	t.Run("CreateGroupInviteNotification", func(t *testing.T) {
		invitedUserID := int64(1)
		inviterUserID := int64(2)
		groupID := int64(3)
		groupName := "Test Group"
		inviterUsername := "testuser"

		payloadBytes, _ := json.Marshal(map[string]string{
			"inviter_id":   "2",
			"inviter_name": "testuser",
			"group_id":     "3",
			"group_name":   "Test Group",
			"action":       "accept_or_decline",
		})

		expectedNotification := sqlc.Notification{
			ID:             1,
			UserID:         invitedUserID,
			NotifType:      string(GroupInvite),
			SourceService:  "users",
			SourceEntityID: pgtype.Int8{Int64: groupID, Valid: true},
			Seen:           pgtype.Bool{Bool: false, Valid: true},
			NeedsAction:    pgtype.Bool{Bool: true, Valid: true},
			Acted:          pgtype.Bool{Bool: false, Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
			DeletedAt:      pgtype.Timestamptz{Valid: false},
			Payload:        payloadBytes,
			Count:          pgtype.Int4{Int32: 1, Valid: true}, // Add the count field
		}

		mockDB.On("CreateNotification", ctx, mock.AnythingOfType("sqlc.CreateNotificationParams")).Return(expectedNotification, nil)

		err := app.CreateGroupInviteNotification(ctx, invitedUserID, inviterUserID, groupID, groupName, inviterUsername)
		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})

	// Test CreateGroupJoinRequestNotification
	t.Run("CreateGroupJoinRequestNotification", func(t *testing.T) {
		groupOwnerID := int64(1)
		requesterID := int64(2)
		groupID := int64(3)
		groupName := "Test Group"
		requesterUsername := "testuser"

		payloadBytes, _ := json.Marshal(map[string]string{
			"requester_id":   "2",
			"requester_name": "testuser",
			"group_id":       "3",
			"group_name":     "Test Group",
			"action":         "accept_or_decline",
		})

		expectedNotification := sqlc.Notification{
			ID:             1,
			UserID:         groupOwnerID,
			NotifType:      string(GroupJoinRequest),
			SourceService:  "users",
			SourceEntityID: pgtype.Int8{Int64: groupID, Valid: true},
			Seen:           pgtype.Bool{Bool: false, Valid: true},
			NeedsAction:    pgtype.Bool{Bool: true, Valid: true},
			Acted:          pgtype.Bool{Bool: false, Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
			DeletedAt:      pgtype.Timestamptz{Valid: false},
			Payload:        payloadBytes,
			Count:          pgtype.Int4{Int32: 1, Valid: true}, // Add the count field
		}

		mockDB.On("CreateNotification", ctx, mock.AnythingOfType("sqlc.CreateNotificationParams")).Return(expectedNotification, nil)

		err := app.CreateGroupJoinRequestNotification(ctx, groupOwnerID, requesterID, groupID, groupName, requesterUsername)
		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})

	// Test CreateNewEventNotification
	t.Run("CreateNewEventNotification", func(t *testing.T) {
		userID := int64(1)
		eventCreatorID := int64(4)
		groupID := int64(2)
		eventID := int64(3)
		groupName := "Test Group"
		eventTitle := "Test Event"

		payloadBytes, _ := json.Marshal(map[string]string{
			"group_id":    "2",
			"group_name":  "Test Group",
			"event_id":    "3",
			"event_title": "Test Event",
			"action":      "view_event",
		})

		expectedNotification := sqlc.Notification{
			ID:             1,
			UserID:         userID,
			NotifType:      string(NewEvent),
			SourceService:  "posts",
			SourceEntityID: pgtype.Int8{Int64: eventID, Valid: true},
			Seen:           pgtype.Bool{Bool: false, Valid: true},
			NeedsAction:    pgtype.Bool{Bool: false, Valid: true},
			Acted:          pgtype.Bool{Bool: false, Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
			DeletedAt:      pgtype.Timestamptz{Valid: false},
			Payload:        payloadBytes,
			Count:          pgtype.Int4{Int32: 1, Valid: true}, // Add the count field
		}

		mockDB.On("CreateNotification", ctx, mock.AnythingOfType("sqlc.CreateNotificationParams")).Return(expectedNotification, nil)

		err := app.CreateNewEventNotification(ctx, userID, eventCreatorID, groupID, eventID, groupName, eventTitle)
		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})
}

// Test notification queries work
func TestNotificationQueries(t *testing.T) {
	mockDB := new(MockDB)
	app := NewApplicationWithMocks(mockDB)

	ctx := context.Background()
	userID := int64(1)

	// Test GetUserUnreadNotificationsCount
	t.Run("GetUserUnreadNotificationsCount", func(t *testing.T) {
		mockDB.On("GetUserUnreadNotificationsCount", ctx, userID).Return(int64(5), nil)

		count, err := app.GetUserUnreadNotificationsCount(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
		mockDB.AssertExpectations(t)
	})

	// Test MarkNotificationAsRead
	t.Run("MarkNotificationAsRead", func(t *testing.T) {
		notificationID := int64(123)
		mockDB.On("MarkNotificationAsRead", ctx, mock.AnythingOfType("sqlc.MarkNotificationAsReadParams")).Return(nil)

		err := app.MarkNotificationAsRead(ctx, notificationID, userID)
		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})
}

// Test aggregation functionality
func TestAggregationFunctionality(t *testing.T) {
	mockDB := new(MockDB)
	app := NewApplicationWithMocks(mockDB)

	ctx := context.Background()
	userID := int64(1)
	postID := int64(100)
	notifType := PostLike
	sourceService := "posts"
	needsAction := false
	payload := map[string]string{"liker_id": "2", "liker_name": "liker1"}
	title := "Post Liked"
	message := "liker1 liked your post"

	payloadBytes, _ := json.Marshal(payload)

	t.Run("Aggregation enabled - New notification when no existing unread notification", func(t *testing.T) {
		// Expect GetUnreadNotificationByTypeAndEntity to return pgx.ErrNoRows (no existing notification)
		mockDB.On("GetUnreadNotificationByTypeAndEntity", ctx, mock.AnythingOfType("sqlc.GetUnreadNotificationByTypeAndEntityParams")).Return(sqlc.Notification{}, pgx.ErrNoRows)

		// Expect CreateNotification to be called
		expectedNotification := sqlc.Notification{
			ID:             1,
			UserID:         userID,
			NotifType:      string(notifType),
			SourceService:  sourceService,
			SourceEntityID: pgtype.Int8{Int64: postID, Valid: true},
			Seen:           pgtype.Bool{Bool: false, Valid: true},
			NeedsAction:    pgtype.Bool{Bool: needsAction, Valid: true},
			Acted:          pgtype.Bool{Bool: false, Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
			DeletedAt:      pgtype.Timestamptz{Valid: false},
			Payload:        payloadBytes,
			Count:          pgtype.Int4{Int32: 1, Valid: true}, // New notification count is 1
		}

		mockDB.On("CreateNotification", ctx, mock.AnythingOfType("sqlc.CreateNotificationParams")).Return(expectedNotification, nil)

		notification, err := app.CreateNotificationWithAggregation(ctx, userID, notifType, title, message, sourceService, postID, needsAction, payload, true)

		assert.NoError(t, err)
		assert.NotNil(t, notification)
		assert.Equal(t, int32(1), notification.Count) // Should be 1 since no existing notification was found
		mockDB.AssertExpectations(t)
	})

	t.Run("Aggregation enabled - Increment count when existing unread notification exists", func(t *testing.T) {
		// Reset mock for this test
		mockDB.ExpectedCalls = nil
		mockDB.Calls = nil

		// First, expect GetUnreadNotificationByTypeAndEntity to return an existing notification with count=2
		existingNotification := sqlc.Notification{
			ID:             10,
			UserID:         userID,
			NotifType:      string(notifType),
			SourceService:  sourceService,
			SourceEntityID: pgtype.Int8{Int64: postID, Valid: true},
			Seen:           pgtype.Bool{Bool: false, Valid: true},
			NeedsAction:    pgtype.Bool{Bool: needsAction, Valid: true},
			Acted:          pgtype.Bool{Bool: false, Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
			DeletedAt:      pgtype.Timestamptz{Valid: false},
			Payload:        payloadBytes,
			Count:          pgtype.Int4{Int32: 2, Valid: true}, // Existing count is 2
		}

		mockDB.On("GetUnreadNotificationByTypeAndEntity", ctx, mock.AnythingOfType("sqlc.GetUnreadNotificationByTypeAndEntityParams")).Return(existingNotification, nil)

		// Then expect UpdateNotificationCount to be called to increment the count to 3
		mockDB.On("UpdateNotificationCount", ctx, mock.AnythingOfType("sqlc.UpdateNotificationCountParams")).Return(nil)

		// Finally expect GetNotificationByID to fetch the updated notification
		updatedNotification := sqlc.Notification{
			ID:             10,
			UserID:         userID,
			NotifType:      string(notifType),
			SourceService:  sourceService,
			SourceEntityID: pgtype.Int8{Int64: postID, Valid: true},
			Seen:           pgtype.Bool{Bool: false, Valid: true},
			NeedsAction:    pgtype.Bool{Bool: needsAction, Valid: true},
			Acted:          pgtype.Bool{Bool: false, Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
			DeletedAt:      pgtype.Timestamptz{Valid: false},
			Payload:        payloadBytes,
			Count:          pgtype.Int4{Int32: 3, Valid: true}, // Updated count is 3
		}

		mockDB.On("GetNotificationByID", ctx, int64(10)).Return(updatedNotification, nil)

		notification, err := app.CreateNotificationWithAggregation(ctx, userID, notifType, title, message, sourceService, postID, needsAction, payload, true)

		assert.NoError(t, err)
		assert.NotNil(t, notification)
		assert.Equal(t, int32(3), notification.Count) // Should be 3 (existing count 2 + 1)
		mockDB.AssertExpectations(t)
	})
}
