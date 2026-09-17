package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"sechelper-auth-template/api/internal/platform/config"
)

var buildEnvironment = "production"

func main() {
	cfg, err := config.Load(os.Args[1:]...)
	if err != nil {
		fail(err)
	}
	if cfg.App.Environment != buildEnvironment {
		fail(fmt.Errorf("migration binary build environment %q does not match config environment %q", buildEnvironment, cfg.App.Environment))
	}
	dsn := cfg.Bootstrap.DatabaseURL
	if dsn == "" {
		fail(fmt.Errorf("bootstrap.databaseUrl is required"))
	}
	db, err := sql.Open("pgx", dsn)
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
	entries, err := collectMigrations(dir)
	if err != nil {
		fail(err)
	}
	includeExamples := os.Getenv("INCLUDE_EXAMPLE_MIGRATIONS") == "1"
	for _, path := range entries {
		name := filepath.Base(path)
		kind, err := migrationModuleKind(dir, path)
		if err != nil {
			fail(err)
		}
		apply, err := shouldApplyMigration(kind, cfg.App.Environment, includeExamples)
		if err != nil {
			fail(err)
		}
		if !apply {
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

func collectMigrations(dir string) ([]string, error) {
	entries := make([]string, 0)
	seen := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			return nil
		}
		name := entry.Name()
		if previous, exists := seen[name]; exists {
			return fmt.Errorf("duplicate migration version %s: %s and %s", name, previous, path)
		}
		seen[name] = path
		entries = append(entries, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		leftName, rightName := filepath.Base(entries[i]), filepath.Base(entries[j])
		if leftName == rightName {
			return entries[i] < entries[j]
		}
		return leftName < rightName
	})
	return entries, nil
}

func migrationModuleKind(root, path string) (string, error) {
	relativeDirectory, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(filepath.ToSlash(relativeDirectory), "business/") {
		return "production", nil
	}
	profilePath := filepath.Join(filepath.Dir(path), "MODULE_KIND")
	contents, err := os.ReadFile(profilePath)
	if os.IsNotExist(err) {
		return "production", nil
	}
	if err != nil {
		return "", fmt.Errorf("read migration module profile %q: %w", profilePath, err)
	}
	kind := strings.TrimSpace(string(contents))
	if kind != "production" && kind != "example" {
		return "", fmt.Errorf("unsupported migration module kind %q in %s", kind, profilePath)
	}
	return kind, nil
}

func shouldApplyMigration(kind, environment string, includeExamples bool) (bool, error) {
	if kind != "production" && kind != "example" {
		return false, fmt.Errorf("unsupported migration module kind %q", kind)
	}
	if environment == "production" && includeExamples {
		return false, fmt.Errorf("example migrations are forbidden in production")
	}
	return kind != "example" || (environment != "production" && includeExamples), nil
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
