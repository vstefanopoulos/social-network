package main

import (
	"social-network/services/testing/internal/entry"
	"time"
)

func main() {
	entry.Run()
	//sleeping to give time to test to finish printing things
	time.Sleep(time.Second)
}
