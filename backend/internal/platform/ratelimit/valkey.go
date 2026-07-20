package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	platformvalkey "github.com/gabrielalc23/pdv/internal/platform/valkey"
)

const fixedWindowScript = `
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
if ttl < 0 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
  ttl = tonumber(ARGV[1])
end
return {count, ttl}
`

var errValkeyUnavailable = errors.New("valkey rate limiter unavailable")

type fixedWindowClient interface {
	increment(ctx context.Context, key string, windowMillis int64) (count, ttlMillis int64, err error)
}

type valkeyFixedWindowClient struct {
	client *platformvalkey.Client
}

type valkeyLimiter struct {
	client   fixedWindowClient
	fallback Limiter
}

// NewValkeyLimiter creates an atomic fixed-window limiter backed by Valkey.
// If Valkey is unavailable, fallback is used for conservative local limiting.
func NewValkeyLimiter(client *platformvalkey.Client, fallback Limiter) Limiter {
	if client == nil {
		return newValkeyLimiter(nil, fallback)
	}
	return newValkeyLimiter(&valkeyFixedWindowClient{client: client}, fallback)
}

func newValkeyLimiter(client fixedWindowClient, fallback Limiter) Limiter {
	return &valkeyLimiter{client: client, fallback: fallback}
}

var _ Limiter = (*valkeyLimiter)(nil)

func (l *valkeyLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	if limit <= 0 {
		return Result{}, ErrInvalidLimit
	}
	if window <= 0 {
		return Result{}, ErrInvalidWindow
	}

	windowMillis := window.Milliseconds()
	if window%time.Millisecond != 0 {
		windowMillis++
	}

	var (
		count     int64
		ttlMillis int64
		err       error
	)
	if l.client == nil {
		err = errValkeyUnavailable
	} else {
		count, ttlMillis, err = l.client.increment(ctx, key, windowMillis)
	}
	if err != nil {
		return l.allowFallback(ctx, key, limit, window, err)
	}
	if count < 1 || ttlMillis < 0 {
		return l.allowFallback(ctx, key, limit, window, errors.New("invalid valkey rate limit response"))
	}

	remaining := max(int64(limit)-count, 0)

	return Result{
		Allowed:   count <= int64(limit),
		Remaining: int(remaining),
		ResetAt:   time.Now().Add(time.Duration(ttlMillis) * time.Millisecond),
	}, nil
}

func (l *valkeyLimiter) allowFallback(ctx context.Context, key string, limit int, window time.Duration, cause error) (Result, error) {
	slog.Warn("rate limiter degraded to local fallback",
		"component", "rate_limiter",
		"backend", "valkey",
		"reason", "command_error",
		"error_type", fmt.Sprintf("%T", cause))

	if l.fallback == nil {
		return Result{Allowed: false, Remaining: 0, ResetAt: time.Now().Add(window)}, nil
	}
	return l.fallback.Allow(ctx, key, limit, window)
}

func (c *valkeyFixedWindowClient) increment(ctx context.Context, key string, windowMillis int64) (int64, int64, error) {
	cmd := c.client.B().Eval().
		Script(fixedWindowScript).
		Numkeys(1).
		Key(key).
		Arg(strconv.FormatInt(windowMillis, 10)).
		Build()

	result, err := c.client.Do(ctx, cmd)
	if err != nil {
		return 0, 0, err
	}

	values, err := result.AsIntSlice()
	if err != nil {
		return 0, 0, err
	}
	if len(values) != 2 {
		return 0, 0, fmt.Errorf("unexpected valkey rate limit response length: %d", len(values))
	}
	return values[0], values[1], nil
}
