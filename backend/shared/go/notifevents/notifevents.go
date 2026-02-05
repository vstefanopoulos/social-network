package notifevents

import (
	"context"
	"fmt"
	notifpb "social-network/shared/gen-go/notifications"
	"social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper to create and send a notification event
func (e *EventCreator) CreateAndSendNotificationEvent(ctx context.Context, event *notifpb.NotificationEvent) error {
	// Extract metadata
	requestId, ok := ctx.Value(ct.ReqID).(string)
	if !ok {
		tele.Error(ctx, "could not get request id")
		requestId = "unknown"
	}

	//TODO figure out otel traces
	metadata := map[string]string{
		"request_id": requestId,
	}

	// Populate common fields
	event.EventId = uuid.NewString()
	event.OccurredAt = timestamppb.Now()
	event.Metadata = metadata

	// Serialize
	eventBytes, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal notification event: %w", err)
	}

	// Send to Kafka
	err = e.kafClient.Send(ctx, ct.NotificationTopic, eventBytes)
	if err != nil {
		return fmt.Errorf("failed to send notification event: %w", err)
	}

	tele.Info(ctx, "Notification event sent: @1", "eventType", event.EventType)
	return nil
}
