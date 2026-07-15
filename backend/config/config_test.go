package config

import "testing"

func TestLoadUsesDefaultsAndDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/pdv?sslmode=disable")
	t.Setenv("HTTP_ADDRESS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Address != ":8080" {
		t.Fatalf("expected default address :8080, got %q", cfg.Address)
	}

	if cfg.DatabaseURL == "" {
		t.Fatalf("expected database url to be set")
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatalf("expected error when DATABASE_URL is missing")
	}
}
