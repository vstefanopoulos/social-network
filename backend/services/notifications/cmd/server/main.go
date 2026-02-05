package main

import (
	"fmt"
	"social-network/services/notifications/internal/entry"
)

func main() {
	err := entry.Run()
	if err != nil {
		fmt.Println(err)
	}
}
