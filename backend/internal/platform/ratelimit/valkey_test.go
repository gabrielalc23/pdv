package ratelimit

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type stubFixedWindowClient struct {
	count        int64
	ttlMillis    int64
	err          error
	key          string
	windowMillis int64
}

func (s *stubFixedWindowClient) increment(_ context.Context, key string, windowMillis int64) (int64, int64, error) {
	s.key = key
	s.windowMillis = windowMillis
	return s.count, s.ttlMillis, s.err
}

func TestValkeyLimiterAllowedResult(t *testing.T) {
	client := &stubFixedWindowClient{count: 2, ttlMillis: 30_000}
	limiter := newValkeyLimiter(client, NewFallbackLimiter(10))
	before := time.Now()

	result, err := limiter.Allow(context.Background(), "caller-supplied-key", 5, time.Minute)
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if !result.Allowed {
		t.Fatal("expected request to be allowed")
	}
	if result.Remaining != 3 {
		t.Fatalf("expected 3 remaining, got %d", result.Remaining)
	}
	if client.key != "caller-supplied-key" {
		t.Fatalf("expected caller key unchanged, got %q", client.key)
	}
	if client.windowMillis != 60_000 {
		t.Fatalf("expected 60000ms window, got %d", client.windowMillis)
	}
	if result.ResetAt.Before(before.Add(29*time.Second)) || result.ResetAt.After(time.Now().Add(31*time.Second)) {
		t.Fatalf("unexpected ResetAt: %v", result.ResetAt)
	}
}

func TestValkeyLimiterDeniedResultSupportsRetryAfter(t *testing.T) {
	client := &stubFixedWindowClient{count: 6, ttlMillis: 2_500}
	limiter := newValkeyLimiter(client, NewFallbackLimiter(10))

	result, err := limiter.Allow(context.Background(), "limited-key", 5, time.Minute)
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected request to be denied")
	}
	if result.Remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", result.Remaining)
	}
	retryAfter := time.Until(result.ResetAt)
	if retryAfter <= 0 || retryAfter > 3*time.Second {
		t.Fatalf("expected positive Retry-After duration, got %v", retryAfter)
	}
}

func TestValkeyLimiterFallsBackOnError(t *testing.T) {
	client := &stubFixedWindowClient{err: errors.New("valkey unavailable")}
	limiter := newValkeyLimiter(client, NewFallbackLimiter(10))
	ctx := context.Background()

	first, err := limiter.Allow(ctx, "fallback-key", 1, time.Minute)
	if err != nil {
		t.Fatalf("first Allow: %v", err)
	}
	if !first.Allowed || first.Remaining != 0 {
		t.Fatalf("unexpected first fallback result: %+v", first)
	}

	second, err := limiter.Allow(ctx, "fallback-key", 1, time.Minute)
	if err != nil {
		t.Fatalf("second Allow: %v", err)
	}
	if second.Allowed || second.Remaining != 0 {
		t.Fatalf("unexpected second fallback result: %+v", second)
	}
}

func TestValkeyLimiterWithoutFallbackDeniesOnError(t *testing.T) {
	limiter := newValkeyLimiter(&stubFixedWindowClient{err: errors.New("valkey unavailable")}, nil)

	result, err := limiter.Allow(context.Background(), "key", 5, time.Minute)
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}
	if result.Allowed || result.Remaining != 0 || result.ResetAt.IsZero() {
		t.Fatalf("expected conservative denial, got %+v", result)
	}
}

func TestFixedWindowScriptIsAtomicAndDoesNotScan(t *testing.T) {
	if !strings.Contains(fixedWindowScript, `redis.call("INCR", KEYS[1])`) ||
		!strings.Contains(fixedWindowScript, `redis.call("PEXPIRE", KEYS[1], ARGV[1])`) ||
		!strings.Contains(fixedWindowScript, `redis.call("PTTL", KEYS[1])`) {
		t.Fatalf("expected atomic fixed-window operations, got %q", fixedWindowScript)
	}
	if strings.Contains(strings.ToUpper(fixedWindowScript), "SCAN") {
		t.Fatalf("script must not use SCAN: %q", fixedWindowScript)
	}
}

func TestValkeyDegradationLogExcludesSensitiveValues(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	secret := "sensitive-raw-key"
	limiter := newValkeyLimiter(&stubFixedWindowClient{err: errors.New("sensitive-raw-error")}, NewFallbackLimiter(10))
	_, err := limiter.Allow(context.Background(), secret, 5, time.Minute)
	if err != nil {
		t.Fatalf("Allow: %v", err)
	}

	entry := logs.String()
	if strings.Contains(entry, secret) || strings.Contains(entry, "sensitive-raw-error") {
		t.Fatalf("degradation log exposed a sensitive raw value: %s", entry)
	}
	if !strings.Contains(entry, `"component":"rate_limiter"`) || !strings.Contains(entry, `"backend":"valkey"`) {
		t.Fatalf("expected structured degradation fields, got %s", entry)
	}
}
