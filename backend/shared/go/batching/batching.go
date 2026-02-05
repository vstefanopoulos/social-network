package batching

import (
	"context"
	"errors"
	"fmt"

	"sync"
	"time"
)

// TODO when writing in error channel have a timer and then panic
// TODO if bucket is emtpy don't call handler (error case)
// Close error chan when finished, and on context cancellation
// error channel timeout on all err channel operations
// TODO when closing batcher, close the error channel and print error log if there was an error, if the input channel closes then print an info

var ErrBatcherTerminated = errors.New("batcher is terminating")

// this batcher is pretty cool, needs testing and edge case finding.
// give it a handler function that takes in a slice of a type T, and it returns an input channel and an error channel. Put things into the input channel you want to be buffered and listened to the error channel for errors.
// this batcher will accumulate messages, but in the case that messages arrive at a slower pace than the batching interval, then all messages will leave immediately instead of being delayed to be accumulated. So it only batches, and delays, messages only when it needs to
// highly recommended that the input buffer is not 0, make it above 1000 if you're unsure
func Batcher[T any](ctx context.Context, handler func([]T) error, poolingDuration time.Duration, inputBuffer int) (chan<- T, <-chan error) {
	inputChannel := make(chan T, inputBuffer)
	errorChan := make(chan error)

	go func() {
		timer := time.NewTimer(time.Hour)
		timer.Stop()
		timerOn := false

		lastFlushTimeStamp := time.Now().Add(-time.Hour)

		messageBucket := []T{}

		var err error
		for {
			select {
			case message, ok := <-inputChannel:
				// tele.Info(ctx, "receive")
				if !ok {
					//sending message
					err = handler(messageBucket)
					if err != nil {
						errorChan <- err
					}

					timer.Reset(time.Second * 5)
					select {
					case errorChan <- errors.New("channel closed"):
					case <-timer.C:
					}
					return
				}
				// message arrived, adding it to bucket
				messageBucket = append(messageBucket, message)
				if timerOn {
					// timer is already activated, therefore we just gather messages until the timer fires
					continue
				}
				//check if more than 'poolingDuration' amount of time has passed after the last flush
				if time.Since(lastFlushTimeStamp) <= poolingDuration {
					//new message came too soon, so we'll just start the timer and wait it to ring before we flush
					timer.Reset(poolingDuration - time.Since(lastFlushTimeStamp))
					timerOn = true
					continue
				}

				//sending message
				err = handler(messageBucket)
				if err != nil {
					errorChan <- err
				}

				//clear the bucket
				messageBucket = messageBucket[:0]

				//record time of flush, so that when next message comes too soon, we set a timer for the remaining duration
				lastFlushTimeStamp = time.Now()

			case <-timer.C:
				// tele.Info(ctx, "timer expired")
				timerOn = false
				if len(messageBucket) == 0 {
					// bucket is empty, therefore no need to send anything
					continue
				}

				//sending message
				// tele.Info(ctx, "flushed timer")
				err = handler(messageBucket)
				if err != nil {
					errorChan <- err
				}

				//clearing bucket
				messageBucket = messageBucket[:0]
				lastFlushTimeStamp = time.Now()

			case <-ctx.Done():
				timer.Stop()

				//sending message
				err = handler(messageBucket)
				if err != nil {
					errorChan <- err
				}

				return
			}
		}
	}()

	return inputChannel, errorChan
}

func BatcherTest() {
	ctx := context.Background()
	total := 0

	handler := func(str []string) error {
		total += len(str)
		fmt.Println("batch:", len(str), "total:", total)
		time.Sleep(time.Millisecond * 50)
		return nil
	}

	inputChan, _ := Batcher(ctx, handler, time.Millisecond*time.Duration(150), 100)

	wg := sync.WaitGroup{}
	for range 10 {
		wg.Go(func() {
			for i := range 10000 {
				time.Sleep(10 * time.Millisecond)
				// tele.Info(ctx, "send")
				inputChan <- fmt.Sprint(i)
			}
		})
	}

	wg.Wait()
	time.Sleep(time.Second)
}
