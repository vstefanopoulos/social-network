package main

import (
	"context"
	"os"
	tele "social-network/shared/go/telemetry"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	var pool *pgxpool.Pool
	var err error

	connStr := os.Getenv("DATABASE_URL")

	for i := range 10 {
		pool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			break
		}
		tele.Warn(ctx, "DB not ready yet @1, @2 ", "attempt", i+1, "error", err.Error())
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		tele.Fatalf("Failed to connect DB: %v", err)
	}
	defer pool.Close()

	tele.Info(ctx, "Connected to users database")

	tele.Info(ctx, "Service ready!")
}
