package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"
)

// migrationLockKey is a stable application-specific key used with PostgreSQL
// advisory locks so that concurrent instances cannot apply migrations twice.
const migrationLockKey int64 = 0x4c50544d

var emptyChecksum = checksum(nil)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT PRIMARY KEY,
    checksum   TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

// Migrate applies every pending .sql file from files in lexical order.
// Each migration is executed atomically alongside its bookkeeping row. An
// advisory lock prevents multiple server instances from racing each other.
// Empty or whitespace-only files are skipped so placeholders are not prematurely recorded.
//
// It returns the versions that were applied during this call.
func Migrate(ctx context.Context, db *sql.DB, files fs.FS) ([]string, error) {
	if db == nil {
		return nil, fmt.Errorf("migrate: database is nil")
	}
	if files == nil {
		return nil, fmt.Errorf("migrate: migration filesystem is nil")
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("migrate: acquire connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return nil, fmt.Errorf("migrate: acquire advisory lock: %w", err)
	}
	defer releaseMigrationLock(conn)

	if _, err := conn.ExecContext(ctx, createMigrationsTable); err != nil {
		return nil, fmt.Errorf("migrate: create schema_migrations table: %w", err)
	}

	applied, err := loadAppliedMigrations(ctx, conn)
	if err != nil {
		return nil, err
	}

	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("migrate: read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	var newlyApplied []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := entry.Name()
		body, err := fs.ReadFile(files, version)
		if err != nil {
			return nil, fmt.Errorf("migrate: read %q: %w", version, err)
		}

		if len(bytes.TrimSpace(body)) == 0 {
			continue
		}

		checksum := checksum(body)
		if previous, ok := applied[version]; ok {
			if previous == emptyChecksum && checksum != emptyChecksum {
				// File was previously recorded while empty; apply migration and update recorded checksum.
				if err := applyMigration(ctx, conn, version, string(body), checksum); err != nil {
					return nil, err
				}
				newlyApplied = append(newlyApplied, version)
				continue
			}
			if previous != checksum {
				return nil, fmt.Errorf(
					"migrate: %q was modified after it was applied (recorded checksum %s, current checksum %s)",
					version, previous, checksum,
				)
			}
			continue
		}

		if err := applyMigration(ctx, conn, version, string(body), checksum); err != nil {
			return nil, err
		}
		newlyApplied = append(newlyApplied, version)
	}

	return newlyApplied, nil
}

func releaseMigrationLock(conn *sql.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", migrationLockKey)
}

func loadAppliedMigrations(ctx context.Context, conn *sql.Conn) (map[string]string, error) {
	rows, err := conn.QueryContext(ctx, "SELECT version, checksum FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("migrate: load applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]string)
	for rows.Next() {
		var version, checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, fmt.Errorf("migrate: scan applied migration: %w", err)
		}
		applied[version] = checksum
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("migrate: iterate applied migrations: %w", err)
	}
	return applied, nil
}

func applyMigration(ctx context.Context, conn *sql.Conn, version, body, checksum string) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("migrate: begin %q: %w", version, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, body); err != nil {
		return fmt.Errorf("migrate: apply %q: %w", version, err)
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO schema_migrations (version, checksum, applied_at)
         VALUES ($1, $2, CURRENT_TIMESTAMP)
         ON CONFLICT (version) DO UPDATE
         SET checksum = EXCLUDED.checksum, applied_at = EXCLUDED.applied_at`,
		version,
		checksum,
	); err != nil {
		return fmt.Errorf("migrate: record %q: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migrate: commit %q: %w", version, err)
	}
	return nil
}

func checksum(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
