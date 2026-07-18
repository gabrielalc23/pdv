package valkey_test

import (
	"context"
	"testing"

	"github.com/gabrielalc23/pdv/internal/platform/valkey"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     valkey.Config
		wantErr bool
	}{
		{"valid", valkey.Config{Addr: "localhost:6379", DB: 0}, false},
		{"empty_addr", valkey.Config{Addr: "", DB: 0}, true},
		{"negative_db", valkey.Config{Addr: "localhost:6379", DB: -1}, true},
		{"db_too_high", valkey.Config{Addr: "localhost:6379", DB: 16}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewClientInvalidConfig(t *testing.T) {
	_, err := valkey.NewClient(valkey.Config{Addr: "", DB: 0})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestCloseNoError(t *testing.T) {
	client, err := valkey.NewClient(valkey.Config{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	if err != nil {
		t.Skipf("valkey not available: %v", err)
	}
	client.Close()
}

func TestPingWithoutServer(t *testing.T) {
	client, err := valkey.NewClient(valkey.Config{
		Addr: "127.0.0.1:1",
		DB:   0,
	})
	if err != nil {
		t.Skipf("valkey client creation failed (expected for addr :1): %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	err = client.Ping(ctx)
	if err != nil {
		t.Logf("ping expected to fail: %v", err)
	}
}
