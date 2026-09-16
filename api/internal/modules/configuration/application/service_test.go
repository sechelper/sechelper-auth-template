package application

import (
	"context"
	"database/sql"
	"errors"
	"sechelper-auth-template/api/internal/modules/configuration/domain"
	"testing"
	"time"
)

type testProtector struct{}

func (testProtector) Encrypt(value string) (string, error) { return "enc:" + value, nil }
func (testProtector) Decrypt(value string) (string, error) {
	if len(value) < 4 || value[:4] != "enc:" {
		return "", errors.New("invalid ciphertext")
	}
	return value[4:], nil
}

type testRepository struct {
	values  map[string]string
	entries map[string]domain.Entry
}

func newTestRepository() *testRepository {
	return &testRepository{values: map[string]string{}, entries: map[string]domain.Entry{}}
}
func (r *testRepository) List(context.Context) ([]domain.Entry, error) {
	result := make([]domain.Entry, 0, len(r.entries))
	for _, value := range r.entries {
		result = append(result, value)
	}
	return result, nil
}
func (r *testRepository) Values(context.Context) (map[string]string, error) { return r.values, nil }
func (r *testRepository) Get(_ context.Context, key string) (domain.Value, error) {
	entry, ok := r.entries[key]
	if !ok {
		return domain.Value{}, sql.ErrNoRows
	}
	return domain.Value{Entry: entry, Value: r.values[key]}, nil
}
func (r *testRepository) Upsert(_ context.Context, key, description string, secret bool, ciphertext, actor string) (domain.Entry, error) {
	entry := domain.Entry{Key: key, Description: description, IsSecret: secret, Version: 1, UpdatedBy: actor, UpdatedAt: time.Now().UTC()}
	if previous, ok := r.entries[key]; ok {
		entry.Version = previous.Version + 1
	}
	r.entries[key], r.values[key] = entry, ciphertext
	return entry, nil
}
func (r *testRepository) Delete(_ context.Context, key string) error {
	delete(r.entries, key)
	delete(r.values, key)
	return nil
}

func TestUpsertEncryptsAndProviderReadsOnlyRequestedValue(t *testing.T) {
	repository := newTestRepository()
	service := NewService(repository, testProtector{})
	if _, err := service.Upsert(context.Background(), "REDIS_URL", "cache", true, "redis://secret", "admin-1"); err != nil {
		t.Fatal(err)
	}
	if got := repository.values["REDIS_URL"]; got != "enc:redis://secret" {
		t.Fatalf("stored value = %q, want encrypted value", got)
	}
	value, ok := service.GetValue(context.Background(), "REDIS_URL")
	if !ok || value != "redis://secret" {
		t.Fatalf("provider returned (%q, %v)", value, ok)
	}
	if _, ok := service.GetValue(context.Background(), "MISSING"); ok {
		t.Fatal("missing configuration unexpectedly resolved")
	}
}

func TestUpsertRejectsNonEnvironmentKeysAndBlankValues(t *testing.T) {
	service := NewService(newTestRepository(), testProtector{})
	for _, test := range []struct{ key, value string }{{"redis_url", "x"}, {"REDIS-URL", "x"}, {"REDIS_URL", " "}} {
		if _, err := service.Upsert(context.Background(), test.key, "", true, test.value, "admin"); err == nil {
			t.Fatalf("Upsert(%q, %q) unexpectedly succeeded", test.key, test.value)
		}
	}
}
