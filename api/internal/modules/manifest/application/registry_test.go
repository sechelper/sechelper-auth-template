package application

import (
	"sechelper-auth-template/api/internal/modules/manifest/domain"
	"testing"
)

func TestRegistryCanonicalizesManifestBeforeHashing(t *testing.T) {
	r := NewRegistry("demo")
	if err := r.Register(domain.Permission{
		Code: " demo:read ", Name: " Read ", Description: " Description ",
		APIs: []domain.API{{Method: " get ", Path: " /orders "}},
	}); err != nil {
		t.Fatal(err)
	}
	snapshot, hash, err := r.Snapshot(1)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Permissions[0].RiskLevel != "normal" || snapshot.Permissions[0].APIs[0].Method != "GET" {
		t.Fatalf("snapshot was not canonicalized: %+v", snapshot.Permissions[0])
	}
	if len(hash) != len("sha256:")+64 || hash[:7] != "sha256:" {
		t.Fatalf("hash = %q, want sha256-prefixed 64-digit digest", hash)
	}
}

func TestRegistryRejectsConflictingPermission(t *testing.T) {
	r := NewRegistry("demo")
	p := domain.Permission{Code: "demo:read", Name: "Read"}
	if err := r.Register(p); err != nil {
		t.Fatal(err)
	}
	p.Name = "Different"
	if err := r.Register(p); err == nil {
		t.Fatal("expected conflict")
	}
}

func TestRegistryAllowsOnlyDefinedRiskLevels(t *testing.T) {
	for _, riskLevel := range []string{"normal", "privileged", "critical", "PRIVILEGED"} {
		r := NewRegistry("demo")
		if err := r.Register(domain.Permission{Code: "demo:read", Name: "Read", RiskLevel: riskLevel}); err != nil {
			t.Fatalf("risk level %q was rejected: %v", riskLevel, err)
		}
	}

	for _, riskLevel := range []string{"high", "low", "custom"} {
		r := NewRegistry("demo")
		if err := r.Register(domain.Permission{Code: "demo:read", Name: "Read", RiskLevel: riskLevel}); err == nil {
			t.Fatalf("risk level %q was accepted", riskLevel)
		}
	}
}
