package main

import "testing"

func TestHealth(t *testing.T) {
	if got := NewApp().Health(); got != "ready" {
		t.Fatalf("Health() = %q, want %q", got, "ready")
	}
}
