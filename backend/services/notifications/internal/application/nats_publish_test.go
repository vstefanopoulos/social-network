package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPublishNotificationToNATS tests the publishNotificationToNATS method
func TestPublishNotificationToNATS(t *testing.T) {
	// Connect to NATS server
	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		// Skip test if NATS server is not available
		t.Skip("NATS server not available, skipping test")
	}
	defer nc.Close()

	app := &Application{
		NatsConn: nc,
	}

	// Create a test notification
	testNotification := &Notification{
		ID:             1,
		UserID:         123,
		Type:           PostLike,
		Title:          "Test Notification",
		Message:        "This is a test notification",
		SourceService:  "posts",
		SourceEntityID: 456,
		Seen:           false,
		NeedsAction:    false,
		Acted:          false,
		Count:          1,
		CreatedAt:      time.Now(),
		Payload: map[string]string{
			"test_key": "test_value",
		},
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

	// Publish the notification to NATS
	ctx := context.Background()
	err = app.publishNotificationToNATS(ctx, testNotification)
	require.NoError(t, err)

	// Wait for the notification to be received
	select {
	case receivedNotification := <-received:
		assert.Equal(t, testNotification.ID, receivedNotification.ID)
		assert.Equal(t, testNotification.UserID, receivedNotification.UserID)
		assert.Equal(t, testNotification.Type, receivedNotification.Type)
		assert.Equal(t, testNotification.Title, receivedNotification.Title)
		assert.Equal(t, testNotification.Message, receivedNotification.Message)
		assert.Equal(t, testNotification.SourceService, receivedNotification.SourceService)
		assert.Equal(t, testNotification.SourceEntityID, receivedNotification.SourceEntityID)
		assert.Equal(t, testNotification.Payload, receivedNotification.Payload)
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for NATS message")
	}
}

// TestPublishNotificationToNATSWithNilConnection tests the method when NATS connection is nil
func TestPublishNotificationToNATSWithNilConnection(t *testing.T) {
	app := &Application{
		NatsConn: nil,
	}

	testNotification := &Notification{
		ID:     1,
		UserID: 123,
		Type:   PostLike,
	}

	// This should not panic and should return nil
	ctx := context.Background()
	err := app.publishNotificationToNATS(ctx, testNotification)
	assert.NoError(t, err)
}
