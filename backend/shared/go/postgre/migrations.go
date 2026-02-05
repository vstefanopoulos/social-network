package postgresql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Run migrations from a given path
func run(dbUrl string, migrationsPath string) error {
	fmt.Printf("db url: %s migrations path %s", dbUrl, migrationsPath)
	sqlDB, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return fmt.Errorf("failed to open DB for migrations: %w", err)
	}
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}

	var dbName string
	if err := sqlDB.QueryRow("select current_database()").Scan(&dbName); err != nil {
		return err
	}
	fmt.Println("Running migrations on database:", dbName)

	// Check if database is dirty
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}
	if dirty {
		log.Printf("Database is dirty at version %d, forcing version", version)
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("failed to force version: %w", err)
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}

// Runs migrations with retries
func RunMigrations(dbUrl string, migrationsPath string) (err error) {
	for range 10 {
		if err = run(dbUrl, migrationsPath); err != nil {
			log.Println("Migration failed, retrying in 2s:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		return nil
	}
	return err
}
