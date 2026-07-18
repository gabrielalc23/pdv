package ratelimit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrRateLimited   = errors.New("rate limit exceeded")
	ErrInvalidLimit  = errors.New("invalid rate limit")
	ErrInvalidWindow = errors.New("invalid rate limit window")
)

type Result struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error)
}

type fallbackEntry struct {
	count   int
	resetAt time.Time
}

type fallbackLimiter struct {
	mu       sync.Mutex
	entries  map[string]*fallbackEntry
	maxItems int
}

func NewFallbackLimiter(maxItems int) Limiter {
	return &fallbackLimiter{
		entries:  make(map[string]*fallbackEntry),
		maxItems: maxItems,
	}
}

var _ Limiter = (*fallbackLimiter)(nil)

func (f *fallbackLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	if limit <= 0 {
		return Result{}, ErrInvalidLimit
	}
	if window <= 0 {
		return Result{}, ErrInvalidWindow
	}

	now := time.Now()

	f.mu.Lock()

	f.evictExpired(now)

	entry, exists := f.entries[key]
	if !exists {
		if len(f.entries) >= f.maxItems {
			f.mu.Unlock()
			slog.Warn("rate limit fallback at capacity, allowing request",
				"key", key, "max_items", f.maxItems)
			return Result{Allowed: true, Remaining: limit - 1, ResetAt: now.Add(window)}, nil
		}
		entry = &fallbackEntry{count: 0, resetAt: now.Add(window)}
		f.entries[key] = entry
	}

	if now.After(entry.resetAt) {
		entry.count = 0
		entry.resetAt = now.Add(window)
	}

	entry.count++
	remaining := limit - entry.count
	if remaining < 0 {
		remaining = 0
	}

	if entry.count > limit {
		f.mu.Unlock()
		return Result{Allowed: false, Remaining: 0, ResetAt: entry.resetAt}, nil
	}

	f.mu.Unlock()
	return Result{Allowed: true, Remaining: remaining, ResetAt: entry.resetAt}, nil
}

func (f *fallbackLimiter) evictExpired(now time.Time) {
	for k, v := range f.entries {
		if now.After(v.resetAt) {
			delete(f.entries, k)
		}
	}
}

func Fingerprint(secret []byte, value string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
