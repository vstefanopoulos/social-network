package main

import (
	"context"
	"log"
	"os"
	postgresql "social-network/shared/go/postgre"
	tele "social-network/shared/go/telemetry"
)

func main() {
	ctx := context.Background()
	tele.Info(ctx, "CHAT SERVICE DB: Running database migrations...")

	if err := postgresql.RunMigrations(os.Getenv("DATABASE_URL"), os.Getenv("MIGRATE_PATH")); err != nil {
		log.Fatal("Migration failed", err)
	}

	tele.Info(ctx, "✅ CHAT SERVICE DB: Migrations completed successfully.")
	os.Exit(0)
}
