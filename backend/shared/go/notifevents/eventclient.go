package notifevents

import (
	"context"
	"social-network/shared/go/ct"
	"social-network/shared/go/kafgo"
)

type EventProducer interface {
	Send(ctx context.Context, topic ct.KafkaTopic, payload ...[]byte) error
}

type EventCreator struct {
	kafClient EventProducer
}

func NewEventProducer(kafClient *kafgo.KafkaProducer) *EventCreator {
	return &EventCreator{kafClient: kafClient}
}
