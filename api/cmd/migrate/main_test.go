package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCollectMigrationsIncludesBusinessDirectoriesInVersionOrder(t *testing.T) {
	dir := t.TempDir()
	writeMigration(t, filepath.Join(dir, "003_framework.sql"))
	writeMigration(t, filepath.Join(dir, "business", "orders", "002_orders.sql"))
	writeMigration(t, filepath.Join(dir, "001_framework.sql"))

	paths, err := collectMigrations(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(paths))
	for _, path := range paths {
		names = append(names, filepath.Base(path))
	}
	want := []string{"001_framework.sql", "002_orders.sql", "003_framework.sql"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("unexpected migration order: got %v want %v", names, want)
	}
}

func TestCollectMigrationsRejectsDuplicateVersions(t *testing.T) {
	dir := t.TempDir()
	writeMigration(t, filepath.Join(dir, "001_duplicate.sql"))
	writeMigration(t, filepath.Join(dir, "business", "orders", "001_duplicate.sql"))

	if _, err := collectMigrations(dir); err == nil {
		t.Fatal("expected duplicate migration version error")
	}
}

func TestMigrationModuleKindDefaultsAndOptInPolicy(t *testing.T) {
	dir := t.TempDir()
	framework := filepath.Join(dir, "001_initial.sql")
	orders := filepath.Join(dir, "business", "orders", "002_orders.sql")
	writeMigration(t, framework)
	writeMigration(t, orders)

	if kind, err := migrationModuleKind(dir, framework); err != nil || kind != "production" {
		t.Fatalf("framework migration kind = %q, err = %v", kind, err)
	}
	if kind, err := migrationModuleKind(dir, orders); err != nil || kind != "production" {
		t.Fatalf("unmarked business migration kind = %q, err = %v", kind, err)
	}
	profile := filepath.Join(filepath.Dir(orders), "MODULE_KIND")
	if err := os.WriteFile(profile, []byte("example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if kind, err := migrationModuleKind(dir, orders); err != nil || kind != "example" {
		t.Fatalf("marked business migration kind = %q, err = %v", kind, err)
	}
	if apply, err := shouldApplyMigration("example", "test", true); err != nil || !apply {
		t.Fatalf("test example migration should apply: apply=%t err=%v", apply, err)
	}
	if apply, err := shouldApplyMigration("example", "production", false); err != nil || apply {
		t.Fatalf("production example migration should be excluded: apply=%t err=%v", apply, err)
	}
	if _, err := shouldApplyMigration("example", "production", true); err == nil {
		t.Fatal("production must reject an explicit example-migration setting")
	}
}

func writeMigration(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("SELECT 1;\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}
