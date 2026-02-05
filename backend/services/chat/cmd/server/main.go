package main

import (
	"fmt"
	"social-network/services/chat/internal/entry"
)

func main() {
	err := entry.Run()
	if err != nil {
		fmt.Println(err)
	}
}
