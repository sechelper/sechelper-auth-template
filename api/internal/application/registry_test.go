package application

import "testing"

type testModule string

func (m testModule) Name() string { return string(m) }

func TestRegistryRejectsDuplicateNames(t *testing.T) {
	r := NewRegistry()
	if err := r.Add(testModule("customer")); err != nil {
		t.Fatal(err)
	}
	if err := r.Add(testModule("customer")); err == nil {
		t.Fatal("expected duplicate module error")
	}
}
