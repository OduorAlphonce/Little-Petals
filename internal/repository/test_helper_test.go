package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"little-petals/internal/database"
	"little-petals/web/migrations"
)

const defaultTestConnStr = "postgres://postgres:postgres@localhost:5432/littlepetals?sslmode=disable"

// setupTestDB connects to the PostgreSQL test database, applies all schema migrations,
// and truncates tables for test isolation.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	connStr := os.Getenv("TEST_DB_URL")
	if connStr == "" {
		connStr = os.Getenv("DB_CONN_STR")
	}
	if connStr == "" {
		connStr = defaultTestConnStr
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Skipf("skipping PostgreSQL integration test: open connection failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Skipf("skipping PostgreSQL integration test: cannot connect to database at %s: %v", connStr, err)
	}

	// Apply embedded migrations to ensure schema is fully up to date.
	if _, err := database.Migrate(ctx, db, migrations.Files); err != nil {
		_ = db.Close()
		t.Fatalf("setupTestDB: failed to run migrations: %v", err)
	}

	// Truncate tables for clean test state.
	cleanTables(t, db)

	t.Cleanup(func() {
		cleanTables(t, db)
		_ = db.Close()
	})

	return db
}

func cleanTables(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, "TRUNCATE TABLE posts, users CASCADE;")
	if err != nil {
		t.Logf("cleanTables: truncate warning: %v", err)
	}
}
