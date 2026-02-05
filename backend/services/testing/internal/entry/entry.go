package entry

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"

	// chattesting "social-network/services/testing/internal/chat_testing"
	chattesting "social-network/services/testing/internal/chat_testing"
	"social-network/services/testing/internal/configs"
	gateway_test "social-network/services/testing/internal/gateway_testing"
	kafkatester "social-network/services/testing/internal/kafka"
	users_test "social-network/services/testing/internal/users_testing"
	tele "social-network/shared/go/telemetry"
	"sync"
	"syscall"
	"time"
)

var cfgs configs.Configs

func catchPanic(ctx context.Context, testName string) {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		tele.Error(ctx, "PANIC in @1: @2. Stack: @3", "test", testName, "panic", fmt.Sprint(r), "stack", stack)
	}
}

func Run() {
	fmt.Println("start run")
	cfgs = configs.GetConfigs()
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}

	for range 1 {
		wg.Go(func() {
			defer catchPanic(ctx, "users_test")
			if err := users_test.StartTest(ctx, cfgs); err != nil {
				tele.Fatal("!!!!!!!!!!!!!!!!!!!!!!!!!!!!! ERROR WTF !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!" + err.Error())
			}
		})

		wg.Go(func() {
			defer catchPanic(ctx, "gateway_test")
			gateway_test.StartTest(ctx, cfgs)
		})

		wg.Go(func() {
			defer catchPanic(ctx, "kafkatester")
			kafkatester.TestKafka()
		})

		wg.Go(func() {
			defer catchPanic(ctx, "chattesting")
			chattesting.StartTest(ctx, cfgs)
		})

		time.Sleep(time.Millisecond * 2000)
	}

	wg.Wait()
	stopSignal()
	fmt.Println("end run")
}
