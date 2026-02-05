package main

import (
	"social-network/services/users/internal/entry"

	_ "github.com/lib/pq"
)

func main() {
	entry.Run()
}
