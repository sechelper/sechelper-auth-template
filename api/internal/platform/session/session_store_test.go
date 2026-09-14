package session

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLoginTransactionIsSingleUse(t *testing.T) {
	store := NewMemoryStore()
	value := LoginTransaction{State: "state-1", Nonce: "nonce", Verifier: "verifier", ExpiresAt: time.Now().Add(time.Minute)}
	if err := store.SaveLoginTransaction(context.Background(), value); err != nil {
		t.Fatal(err)
	}
	got, err := store.ConsumeLoginTransaction(context.Background(), value.State)
	if err != nil || got.Nonce != value.Nonce {
		t.Fatalf("ConsumeLoginTransaction() = (%+v, %v)", got, err)
	}
	if _, err := store.ConsumeLoginTransaction(context.Background(), value.State); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second ConsumeLoginTransaction() error = %v, want ErrNotFound", err)
	}
}

func TestSessionUpdateRejectsStaleVersion(t *testing.T) {
	store := NewMemoryStore()
	if err := store.Create(context.Background(), Session{ID: "session-1", Subject: "user", ExpiresAt: time.Now().Add(time.Minute)}); err != nil {
		t.Fatal(err)
	}
	first, err := store.Get(context.Background(), "session-1")
	if err != nil {
		t.Fatal(err)
	}
	first.Email = "first@example.test"
	if err := store.Update(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	first.Email = "stale@example.test"
	if err := store.Update(context.Background(), first); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale Update() error = %v, want ErrConflict", err)
	}
}
