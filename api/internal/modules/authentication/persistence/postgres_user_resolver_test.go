package persistence

import (
	"regexp"
	"testing"
)

func TestNewUUIDReturnsRFC4122Version4(t *testing.T) {
	value, err := newUUID()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(value) {
		t.Fatalf("newUUID() = %q, not an RFC 4122 version 4 UUID", value)
	}
}
