package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielalc23/pdv/internal/platform/ratelimit"
)

func TestFirstRequestAllowed(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)

	result, err := limiter.Allow(context.Background(), "test-key", 5, time.Minute)
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if !result.Allowed {
		t.Fatal("expected first request to be allowed")
	}
	if result.Remaining != 4 {
		t.Fatalf("expected 4 remaining, got %d", result.Remaining)
	}
}

func TestLimitReached(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		result, err := limiter.Allow(ctx, "limit-key", 5, time.Minute)
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	result, err := limiter.Allow(ctx, "limit-key", 5, time.Minute)
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected rate limited")
	}
	if result.Remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", result.Remaining)
	}
}

func TestLimitExceeded(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		limiter.Allow(ctx, "exceed-key", 3, time.Minute)
	}

	result, _ := limiter.Allow(ctx, "exceed-key", 3, time.Minute)
	if result.Allowed {
		t.Fatal("expected rate limited after exceeding")
	}
}

func TestRemainingCount(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)
	ctx := context.Background()

	expectedRemaining := []int{4, 3, 2, 1, 0}
	for i, exp := range expectedRemaining {
		result, err := limiter.Allow(ctx, "rem-key", 5, time.Minute)
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if result.Remaining != exp {
			t.Fatalf("request %d: expected %d remaining, got %d", i, exp, result.Remaining)
		}
	}
}

func TestWindowExpiration(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "exp-key", 2, 50*time.Millisecond)
	}

	result, _ := limiter.Allow(ctx, "exp-key", 2, 50*time.Millisecond)
	if result.Allowed {
		t.Log("still rate limited (50ms hasn't passed)")
	}

	<-time.After(60 * time.Millisecond)

	result, err := limiter.Allow(ctx, "exp-key", 2, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("Allow after window: %v", err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed after window expiration")
	}
}

func TestIsolationByKey(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, "key-a", 5, time.Minute)
	}

	result, _ := limiter.Allow(ctx, "key-b", 5, time.Minute)
	if !result.Allowed {
		t.Fatal("key-b should be independent from key-a")
	}
}

func TestInvalidLimit(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)

	_, err := limiter.Allow(context.Background(), "key", 0, time.Minute)
	if err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestInvalidWindow(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)

	_, err := limiter.Allow(context.Background(), "key", 5, 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestFallbackCapacity(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(3)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		result, err := limiter.Allow(ctx, "cap-key-"+string(rune('a'+i)), 5, time.Minute)
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if !result.Allowed {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	result, err := limiter.Allow(ctx, "new-key", 5, time.Minute)
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected fallback to deny when at capacity")
	}
	if result.Remaining != 0 {
		t.Fatalf("expected 0 remaining at capacity, got %d", result.Remaining)
	}
	if result.ResetAt.IsZero() {
		t.Fatal("expected non-zero ResetAt at capacity")
	}
}

func TestFingerprint(t *testing.T) {
	secret := []byte("my-secret-key-32-bytes-long!!!!!!")
	result1 := ratelimit.Fingerprint(secret, "user@example.com")
	result2 := ratelimit.Fingerprint(secret, "user@example.com")

	if result1 != result2 {
		t.Fatal("fingerprint should be deterministic")
	}

	if result1 == "user@example.com" {
		t.Fatal("fingerprint should not expose original value")
	}

	result3 := ratelimit.Fingerprint(secret, "other@example.com")
	if result1 == result3 {
		t.Fatal("different inputs should produce different fingerprints")
	}
}

func TestFingerprintEmptySecret(t *testing.T) {
	result := ratelimit.Fingerprint([]byte{}, "test@example.com")
	if result == "" {
		t.Fatal("fingerprint should not be empty even with empty secret")
	}
}

func TestConcurrentAccess(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(1000)
	ctx := context.Background()

	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_, err := limiter.Allow(ctx, "concurrent-key", 1000, time.Minute)
				if err != nil {
					t.Logf("concurrent error: %v", err)
				}
			}
			done <- true
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestResetAt(t *testing.T) {
	limiter := ratelimit.NewFallbackLimiter(100)
	ctx := context.Background()

	result, err := limiter.Allow(ctx, "reset-key", 5, time.Minute)
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}

	if result.ResetAt.IsZero() {
		t.Fatal("expected non-zero ResetAt")
	}

	if result.ResetAt.Before(time.Now()) {
		t.Fatal("ResetAt should be in the future")
	}
}
