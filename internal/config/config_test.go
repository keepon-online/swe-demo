package config

import "testing"

func TestLoadUsesDefaultPortWhenEnvMissing(t *testing.T) {
	t.Setenv("PORT", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
}
