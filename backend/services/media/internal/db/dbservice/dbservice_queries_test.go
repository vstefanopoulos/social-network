package dbservice

import (
	"context"
	"os"
	"path/filepath"
	ct "social-network/shared/go/ct"
	postgresql "social-network/shared/go/postgre"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const testDBURL = "postgres://postgres:secret@localhost:5437/social_media_test?sslmode=disable"

func setupTestDB(t *testing.T) (*Queries, func()) {
	ctx := context.Background()

	// Run migrations
	migrationsPath := filepath.Join("..", "migrations")
	dir, _ := os.Getwd()
	t.Log("Applying migrations from:", dir)
	err := postgresql.RunMigrations(
		testDBURL,
		migrationsPath,
	)
	require.NoError(t, err)

	// Connect via pgx
	pool, err := pgxpool.New(ctx, testDBURL)
	require.NoError(t, err)

	q := NewQuerier(pool)

	teardown := func() {
		pool.Close()
	}

	return q, teardown
}

func TestQuerier(t *testing.T) {
	q, teardown := setupTestDB(t)
	defer teardown()

	t.Run("CreateFile, GetFileById, CreateVariant, GetVariant", func(t *testing.T) {
		ctx := context.Background()
		obj := uuid.NewString()
		// Test data
		file := File{
			Filename:   "test.jpg",
			MimeType:   "image/jpeg",
			SizeBytes:  1024,
			Bucket:     "test-bucket",
			ObjectKey:  obj,
			Visibility: ct.Public, // Adjust to your enum
			Status:     ct.Complete,
		}

		// Create file
		fileId, err := q.CreateFile(ctx, file)
		require.NoError(t, err)
		require.NotZero(t, fileId)

		// Get file
		retrieved, err := q.GetFileById(ctx, fileId)
		require.NoError(t, err)
		require.Equal(t, file.Filename, retrieved.Filename)

		variant := File{
			Id:         fileId,
			Filename:   "test.jpg",
			MimeType:   "image/jpeg",
			SizeBytes:  1024,
			Variant:    ct.ImgMedium,
			Bucket:     "test-bucket",
			ObjectKey:  obj,
			Visibility: ct.Public,
			Status:     ct.Pending,
		}

		_, err = q.CreateVariant(ctx, variant)
		require.NoError(t, err)
		require.NotZero(t, fileId)

		retrieved, err = q.GetVariant(ctx, fileId, ct.ImgMedium)
		require.NoError(t, err)
		require.Equal(t, file.Filename, retrieved.Filename)

		err = q.UpdateVariantStatusAndSize(ctx, fileId, ct.Complete, retrieved.SizeBytes/2)
		require.NoError(t, err)
	})
}
