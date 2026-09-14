package session

import "testing"

func TestTokenProtectorRoundTrip(t *testing.T) {
	protector, err := NewTokenProtector("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := protector.Encrypt("refresh-token")
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "refresh-token" {
		t.Fatal("token must be encrypted")
	}
	decrypted, err := protector.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "refresh-token" {
		t.Fatalf("unexpected decrypted value %q", decrypted)
	}
}
