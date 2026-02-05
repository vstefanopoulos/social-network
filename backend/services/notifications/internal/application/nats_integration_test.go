package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	db "social-network/services/notifications/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestNATSIntegration tests that notifications are properly published to NATS
func TestNATSIntegration(t *testing.T) {
	// Note: This test requires a running NATS server
	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		t.Skip("NATS server not available, skipping test")
	}
	defer nc.Close()

	// Create a mock DB using the existing helper
	mockDB := new(MockDB)

	// Create the application with the NATS connection
	app := &Application{
		DB:       mockDB,
		NatsConn: nc,
	}

	// Subscribe to the notification subject
	received := make(chan *Notification, 1)
	sub, err := nc.Subscribe("ntf.123", func(m *nats.Msg) {
		var notification Notification
		err := json.Unmarshal(m.Data, &notification)
		if err != nil {
			t.Errorf("Failed to unmarshal notification: %v", err)
			return
		}
		received <- &notification
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Set up expectations for the mock
	expectedParams := db.CreateNotificationParams{
		UserID:         123,
		NotifType:      "post_like",
		SourceService:  "posts",
		SourceEntityID: pgtype.Int8{Int64: 456, Valid: true},
		NeedsAction:    pgtype.Bool{Bool: false, Valid: true},
		Acted:          pgtype.Bool{Bool: false, Valid: true},
		Payload:        []byte(`{"liker_id":"789","liker_name":"test_user","post_id":"456","action":"view_post"}`),
		ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
		Count:          pgtype.Int4{Int32: 1, Valid: true},
	}

	expectedDBNotification := db.Notification{
		ID:             1,
		UserID:         123,
		NotifType:      "post_like",
		SourceService:  "posts",
		SourceEntityID: pgtype.Int8{Int64: 456, Valid: true},
		Seen:           pgtype.Bool{Bool: false, Valid: true},
		NeedsAction:    pgtype.Bool{Bool: false, Valid: true},
		Acted:          pgtype.Bool{Bool: false, Valid: true},
		Payload:        []byte(`{"liker_id":"789","liker_name":"test_user","post_id":"456","action":"view_post"}`),
		CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
		Count:          pgtype.Int4{Int32: 1, Valid: true},
	}

	mockDB.On("CreateNotification", mock.Anything, expectedParams).Return(expectedDBNotification, nil)

	// Call the createNotification method
	ctx := context.Background()
	notification, err := app.createNotification(ctx, 123, PostLike, "Test Title", "Test Message", "posts", 456, false, map[string]string{
		"liker_id":   "789",
		"liker_name": "test_user",
		"post_id":    "456",
		"action":     "view_post",
	}, 1)

	require.NoError(t, err)
	assert.NotNil(t, notification)

	// Wait for the notification to be published to NATS
	select {
	case receivedNotification := <-received:
		assert.Equal(t, int64(123), receivedNotification.UserID)
		assert.Equal(t, PostLike, receivedNotification.Type)
		assert.Equal(t, "Test Title", receivedNotification.Title)
		assert.Equal(t, "Test Message", receivedNotification.Message)
		assert.Equal(t, "posts", receivedNotification.SourceService)
		assert.Equal(t, int64(456), receivedNotification.SourceEntityID)
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for NATS message")
	}

	// Verify that the mock was called as expected
	mockDB.AssertExpectations(t)
}

// TestNATSErrorHandling tests that NATS errors are handled gracefully
func TestNATSErrorHandling(t *testing.T) {
	// Create a mock DB using the existing helper
	mockDB := new(MockDB)

	// Create the application with a closed NATS connection to simulate error
	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		t.Skip("NATS server not available, skipping test")
	}
	nc.Close() // Close the connection to force an error

	app := &Application{
		DB:       mockDB,
		NatsConn: nc,
	}

	// Set up expectations for the mock
	expectedParams := db.CreateNotificationParams{
		UserID:         123,
		NotifType:      "post_like",
		SourceService:  "posts",
		SourceEntityID: pgtype.Int8{Int64: 456, Valid: true},
		NeedsAction:    pgtype.Bool{Bool: false, Valid: true},
		Acted:          pgtype.Bool{Bool: false, Valid: true},
		Payload:        []byte(`{"liker_id":"789","liker_name":"test_user","post_id":"456","action":"view_post"}`),
		ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
		Count:          pgtype.Int4{Int32: 1, Valid: true},
	}

	expectedDBNotification := db.Notification{
		ID:             1,
		UserID:         123,
		NotifType:      "post_like",
		SourceService:  "posts",
		SourceEntityID: pgtype.Int8{Int64: 456, Valid: true},
		Seen:           pgtype.Bool{Bool: false, Valid: true},
		NeedsAction:    pgtype.Bool{Bool: false, Valid: true},
		Acted:          pgtype.Bool{Bool: false, Valid: true},
		Payload:        []byte(`{"liker_id":"789","liker_name":"test_user","post_id":"456","action":"view_post"}`),
		CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
		Count:          pgtype.Int4{Int32: 1, Valid: true},
	}

	mockDB.On("CreateNotification", mock.Anything, expectedParams).Return(expectedDBNotification, nil)

	// Call the createNotification method - this should still succeed despite NATS error
	ctx := context.Background()
	notification, err := app.createNotification(ctx, 123, PostLike, "Test Title", "Test Message", "posts", 456, false, map[string]string{
		"liker_id":   "789",
		"liker_name": "test_user",
		"post_id":    "456",
		"action":     "view_post",
	}, 1)

	// The notification creation should succeed even if NATS publishing fails
	require.NoError(t, err)
	assert.NotNil(t, notification)

	// Verify that the mock was called as expected
	mockDB.AssertExpectations(t)
}