package kafgo

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"social-network/shared/go/batching"
	"social-network/shared/go/ct"
	"sort"
	"strconv"
	"sync"
	"time"
)

func FullKafkaTest(size int) error {

	randomTopic := "test_" + fmt.Sprint(rand.IntN(10000))

	messages := []string{}
	for i := range size {
		messages = append(messages, fmt.Sprint(i))
	}

	wg := sync.WaitGroup{}

	wg.Go(func() {
		err := TestKafkaProducer(messages, ct.KafkaTopic(randomTopic))
		if err != nil {
			// tele.Error(context.Background(), "huge kafka test producer @1", "error", err.Error())
			return
		}
	})

	err := TestKafkaConsumer(messages, randomTopic)
	if err != nil {
		// tele.Error(ctx, "failed consumer test @1", "error", err.Error())
		return err
	}
	// tele.Info(context.Background(), "success!")
	return nil
}

func TestKafkaProducer(messages []string, topic ct.KafkaTopic) error {
	ctx := context.Background()

	producer, _, err := NewKafkaProducer([]string{"kafka:9092"})
	if err != nil {
		return fmt.Errorf("failed to create producer: %w", err)
	}
	handler := func(messages [][]byte) error {
		// tele.Info(ctx, "batcher sending, @1 @2", "from", string(messages[0]), "to", string(messages[len(messages)-1]))
		err := producer.Send(ctx, topic, messages...)
		if err != nil {
			return fmt.Errorf("failed to send messages, err: %w", err)
		}
		return nil
	}

	batchInput, errChan := batching.Batcher(ctx, handler, time.Millisecond*100, 1000)

	wg := sync.WaitGroup{}
	wg.Go(func() {
		for _, msg := range messages {

			select {
			case batchInput <- []byte(msg):
			case <-ctx.Done():
				return
			}
			time.Sleep(time.Millisecond * 10)
		}
		ctx.Done()
	})

	wg.Go(func() {
		for err := range errChan {
			if err != nil {
				// tele.Error(ctx, "found error in producer huge test @1", "error", err.Error())
				ctx.Done()
				return
			}
		}
	})

	wg.Wait()
	return nil
}

func TestKafkaConsumer(messages []string, topic string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer, err := NewKafkaConsumer(
		[]string{"kafka:9092"},
		"test",
		ct.KafkaTopic(topic))
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	consumer.
		WithCommitBuffer(50).
		WithDeadmanTimeout(time.Millisecond * 50).
		WithUncommitLimit(100)

	ch, closeAll, err := consumer.StartConsuming(ctx)
	if err != nil {
		// tele.Error(ctx, "failed to start consuming @1", "error", err.Error())
	}
	defer closeAll()

	found := []string{}
	mu := sync.Mutex{}
	loopCtx, loopCancel := context.WithCancel(ctx)
	consumerRoutine := func() {
		// me := rand.IntN(10000)
		for {
			// tele.Info(loopCtx, "@1 start loop", "me", me)
			select {
			case record := <-ch:

				msg := string(record.rec.Value)
				// tele.Info(loopCtx, "@1 [CONSUMER] received @2", "me", me, "msg", msg)

				dur := min(
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
					rand.IntN(2000),
				)
				time.Sleep(time.Millisecond * time.Duration(dur))

				// tele.Info(loopCtx, "@1 [CONSUMER] committing @2 after waiting for @3ms", "me", me, "msg", msg, "time", dur)
				err := record.Commit(loopCtx)
				if err != nil {
					// tele.Error(context.Background(), "consumer huge test @1", "error", err.Error())
					return
				}
				// tele.Info(loopCtx, "@1 [CONSUMER] committed @2", "me", me, "msg", msg)

				mu.Lock()
				// tele.Info(loopCtx, "@1 [CONSUMER] adding @2", "me", me, "msg", msg)
				found = append(found, msg)

				if len(found) == len(messages) {
					mu.Unlock()
					loopCancel()
					return
				}
				mu.Unlock()
			case <-loopCtx.Done():
				// tele.Info(loopCtx, "@1 [CONSUMER] DONE", "me", me)
				return
			}

		}
	}

	wg := sync.WaitGroup{}
	for range 100 {
		wg.Go(func() { consumerRoutine() })
	}

	wg.Wait()

	sort.Slice(found, func(i, j int) bool {
		jVal, _ := strconv.Atoi(found[j])
		iVal, _ := strconv.Atoi(found[i])
		return iVal < jVal
	})

	for i, msg := range found {
		// tele.Info(ctx, "comparing -> @1 and @2", "found", msg, "expected", fmt.Sprint(i))
		if msg != fmt.Sprint(i) {
			return errors.New(fmt.Sprint("final expected matching test failed! at index: ", i, " f:", msg, " is not same as:", fmt.Sprint(i)))
		}
	}

	if len(messages) != len(found) {
		return fmt.Errorf("incorrect expected found count msgs: %d, found: %d", len(messages), len(found))
	}
	// tele.Info(ctx, "Testing success! Waiting before closing.")
	time.Sleep(time.Second * 5)
	// tele.Info(ctx, "CONSUMER ENDED")
	return nil
}
