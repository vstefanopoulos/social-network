package kafgo

import (
	"context"
	"fmt"
	"social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"

	"github.com/twmb/franz-go/pkg/kgo"
)

// How to use this. Create a kafka producer.
// User Send() to send payloads.
// The payload is a byte slice.

type KafkaProducer struct {
	client *kgo.Client
}

// seeds are used for finding the cluster, just give as many kafka ip's you have
//
// Usage:
//
//	   producer, close, err := kafgo.NewKafkaProducer([]string{"localhost:9092"})
//	   if err != nil {
//		   tele.Fatal("wtf")
//	    }
//	    defer close()
//
//		//then use this to send messages to kafka
//	    err := producer.Send(ctx, topic, payload)
//		err != nil{
//			tele.Fatal("wtf2")
//		}
func NewKafkaProducer(seeds []string) (producer *KafkaProducer, close func(), err error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, nil, err
	}
	kfc := &KafkaProducer{
		client: cl,
	}
	return kfc, cl.Close, nil
}

// TODO batch sends instead of doing one by one

// Send sends payload(s) to the specified topic
func (kfc *KafkaProducer) Send(ctx context.Context, topic ct.KafkaTopic, payload ...[]byte) error {
	records := make([]*kgo.Record, len(payload))
	for i, p := range payload {
		records[i] = &kgo.Record{Topic: string(topic), Value: p}
	}
	results := kfc.client.ProduceSync(ctx, records...)
	if results.FirstErr() != nil {
		tele.Error(ctx, "failed to produce: @1", "error", results.FirstErr().Error())
		return fmt.Errorf("failed to produce %w", results.FirstErr())
	}
	return nil
}
