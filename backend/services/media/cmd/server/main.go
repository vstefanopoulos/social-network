package main

import (
	"context"
	"social-network/services/media/internal/entry"
	tele "social-network/shared/go/telemetry"
)

func main() {
	err := entry.Run()
	if err != nil {
		tele.Error(context.Background(), "media main error. @1"+err.Error())
	}
}
