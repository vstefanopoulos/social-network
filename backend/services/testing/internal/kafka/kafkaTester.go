package kafkatester

import (
	"context"
	"social-network/shared/go/kafgo"
	tele "social-network/shared/go/telemetry"
)

func TestKafka() {
	err := kafgo.FullKafkaTest(1000)
	if err != nil {
		tele.Error(context.Background(), "------------  FAIL TEST: err -> Full kafka test failed: @1", "error", err.Error())
		return
	}
	tele.Info(context.Background(), "------------  SUCCESS -> Full kafka test succeeded")
}
