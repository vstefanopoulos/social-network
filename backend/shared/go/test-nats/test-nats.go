package testnats

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

func TestNATS(nc *nats.Conn) error {
	const (
		subject       = "test.subject"
		sendRate      = 10
		sendSeconds   = 1
		totalMessages = sendRate * sendSeconds
	)

	// Generate random payloads
	sent := make([]string, totalMessages)
	for i := range totalMessages {
		sent[i] = fmt.Sprintf("%d", rand.Intn(1000))
	}

	var wg sync.WaitGroup

	var received []string

	// Subscriber
	wg.Go(func() {
		sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
			payload := string(msg.Data)
			fmt.Println("recv:", payload)
			received = append(received, payload)
		})
		if err != nil {
			return
		}
		defer sub.Unsubscribe()

		nc.Flush()
		time.Sleep(time.Duration(sendSeconds)*time.Second + 200*time.Millisecond)
	})

	// Publisher
	wg.Go(func() {
		ticker := time.NewTicker(time.Second / sendRate)
		defer ticker.Stop()

		for i := range totalMessages {
			<-ticker.C
			_ = nc.Publish(subject, []byte(sent[i]))
		}
		nc.Flush()
	})

	wg.Wait()

	if len(sent) != len(received) {
		return fmt.Errorf("message count mismatch: sent=%d received=%d", len(sent), len(received))
	}

	for i := range sent {
		if sent[i] != received[i] {
			return fmt.Errorf("message mismatch at %d: sent=%s received=%s", i, sent[i], received[i])
		}
	}

	fmt.Println("NATS test passed")
	return nil
}
