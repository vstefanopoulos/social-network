package main

import (
	"log"
	"os"

	postgresql "social-network/shared/go/postgre"
)

func main() {
	log.Println("NOTIFICATIONS SERVICE DB: Running database migrations...")
	dbUrl := os.Getenv("DATABASE_URL")
	if err := postgresql.RunMigrations(dbUrl, os.Getenv("MIGRATE_PATH")); err != nil {
		log.Fatal("Migration failed", err)
	}

	log.Println("✅ NOTIFICATIONS SERVICE DB: Migrations completed successfully.")
	os.Exit(0)
}
