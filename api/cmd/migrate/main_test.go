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
	writeMigration(t, filepath.Join(dir, "business", "billing", "002_billing.sql"))
	writeMigration(t, filepath.Join(dir, "001_framework.sql"))

	paths, err := collectMigrations(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(paths))
	for _, path := range paths {
		names = append(names, filepath.Base(path))
	}
	want := []string{"001_framework.sql", "002_billing.sql", "003_framework.sql"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("unexpected migration order: got %v want %v", names, want)
	}
}

func TestCollectMigrationsRejectsDuplicateVersions(t *testing.T) {
	dir := t.TempDir()
	writeMigration(t, filepath.Join(dir, "001_duplicate.sql"))
	writeMigration(t, filepath.Join(dir, "business", "billing", "001_duplicate.sql"))

	if _, err := collectMigrations(dir); err == nil {
		t.Fatal("expected duplicate migration version error")
	}
}

func TestMigrationModuleKindDefaultsAndOptInPolicy(t *testing.T) {
	dir := t.TempDir()
	framework := filepath.Join(dir, "001_initial.sql")
	billing := filepath.Join(dir, "business", "billing", "002_billing.sql")
	writeMigration(t, framework)
	writeMigration(t, billing)

	if kind, err := migrationModuleKind(dir, framework); err != nil || kind != "production" {
		t.Fatalf("framework migration kind = %q, err = %v", kind, err)
	}
	if kind, err := migrationModuleKind(dir, billing); err != nil || kind != "production" {
		t.Fatalf("unmarked business migration kind = %q, err = %v", kind, err)
	}
	profile := filepath.Join(filepath.Dir(billing), "MODULE_KIND")
	if err := os.WriteFile(profile, []byte("example\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if kind, err := migrationModuleKind(dir, billing); err != nil || kind != "example" {
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

func TestMigrationScope(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "framework", path: filepath.Join(dir, "001_authentication.sql"), want: "framework"},
		{name: "configuration", path: filepath.Join(dir, "008_configuration_center.sql"), want: "configuration"},
		{name: "business", path: filepath.Join(dir, "business", "billing", "001_billing.sql"), want: "business_billing"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := migrationScope(dir, test.path)
			if err != nil || got != test.want {
				t.Fatalf("migrationScope() = %q, %v; want %q", got, err, test.want)
			}
		})
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
