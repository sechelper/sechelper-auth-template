package domain

import (
	"context"
	"testing"
)

func TestContextRoundTripsThroughRequestContext(t *testing.T) {
	want := Context{Subject: "identity-subject", PlatformUserUUID: "89cf8f29-9954-470d-b3de-8a37d20c6f44", ApplicationCode: "app-1"}
	got, ok := FromContext(WithContext(context.Background(), want))
	if !ok || got.Subject != want.Subject || got.PlatformUserUUID != want.PlatformUserUUID || got.ApplicationCode != want.ApplicationCode {
		t.Fatalf("FromContext() = (%+v, %t), want (%+v, true)", got, ok, want)
	}
}
