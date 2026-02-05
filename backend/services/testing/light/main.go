package main

import (
	"context"
	chattesting "social-network/services/testing/internal/chat_testing"
	"social-network/services/testing/internal/configs"
)

func main() {
	chattesting.StartTest(context.Background(), configs.Configs{
		UsersGRPCAddr: "localhost:50051",
		ChatGRPCAddr:  "localhost:50053",
	})
}
