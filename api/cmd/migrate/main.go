package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"sechelper-auth-template/api/internal/platform/config"
)

func main() {
	cfg, err := config.Load(os.Args[1:]...)
	if err != nil {
		fail(err)
	}
	db, err := sql.Open("pgx", cfg.Database.URL())
	if err != nil {
		fail(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fail(err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, checksum TEXT NOT NULL, applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		fail(err)
	}
	// Serialize migration runners across replicas. The advisory lock is held
	// for the complete migration process and released by the connection.
	lockConn, err := db.Conn(ctx)
	if err != nil {
		fail(fmt.Errorf("open migration lock connection: %w", err))
	}
	if _, err := lockConn.ExecContext(ctx, `SELECT pg_advisory_lock(hashtextextended('sechelper-auth-template:migrations', 0))`); err != nil {
		_ = lockConn.Close()
		fail(fmt.Errorf("acquire migration lock: %w", err))
	}
	defer func() {
		_, _ = lockConn.ExecContext(context.Background(), `SELECT pg_advisory_unlock(hashtextextended('sechelper-auth-template:migrations', 0))`)
		_ = lockConn.Close()
	}()

	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}
	entries, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		fail(err)
	}
	sort.Strings(entries)
	for _, path := range entries {
		name := filepath.Base(path)
		// Example business migrations are opt-in and never part of the framework schema.
		if name == "002_orders.sql" && os.Getenv("INCLUDE_EXAMPLE_MIGRATIONS") != "1" {
			continue
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			fail(err)
		}
		checksum := fmt.Sprintf("%x", sha256.Sum256(contents))
		var stored string
		err = db.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE version=$1`, name).Scan(&stored)
		if err == nil {
			if stored != checksum {
				fail(fmt.Errorf("migration %s checksum changed", name))
			}
			continue
		}
		if err != sql.ErrNoRows {
			fail(err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			fail(err)
		}
		if _, err = tx.ExecContext(ctx, string(contents)); err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, checksum) VALUES ($1,$2)`, name, checksum)
		}
		if err != nil {
			_ = tx.Rollback()
			fail(fmt.Errorf("apply %s: %w", name, err))
		}
		if err := tx.Commit(); err != nil {
			fail(err)
		}
		fmt.Println("applied", name)
	}
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
