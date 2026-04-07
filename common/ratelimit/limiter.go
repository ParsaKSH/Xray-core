// Package ratelimit provides a token bucket rate limiter for bandwidth control.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements a token bucket rate limiter.
// It is safe for concurrent use.
type Limiter struct {
	mu       sync.Mutex
	rate     float64   // tokens (bytes) per second
	burst    float64   // maximum tokens
	tokens   float64   // current available tokens
	lastTime time.Time // last time tokens were refilled
}

// NewLimiter creates a new rate limiter with the given rate in bytes per second.
// The burst size is set to max(rate, 64KB) to allow reasonable bursting.
// Returns nil if bytesPerSecond is 0 (unlimited).
func NewLimiter(bytesPerSecond uint64) *Limiter {
	if bytesPerSecond == 0 {
		return nil
	}
	rate := float64(bytesPerSecond)
	burst := rate
	if burst < 65536 {
		burst = 65536 // minimum 64KB burst
	}
	return &Limiter{
		rate:     rate,
		burst:    burst,
		tokens:   burst, // start full
		lastTime: time.Now(),
	}
}

// Wait blocks until n bytes are allowed or the context is cancelled.
// If the limiter is nil, it returns immediately (unlimited).
func (l *Limiter) Wait(ctx context.Context, n int) error {
	if l == nil || n <= 0 {
		return nil
	}

	l.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.tokens += elapsed * l.rate
	if l.tokens > l.burst {
		l.tokens = l.burst
	}
	l.lastTime = now

	// Consume tokens (can go negative to "reserve" future capacity)
	l.tokens -= float64(n)
	var waitDuration time.Duration
	if l.tokens < 0 {
		waitDuration = time.Duration(-l.tokens / l.rate * float64(time.Second))
	}
	l.mu.Unlock()

	if waitDuration <= 0 {
		return nil
	}

	t := time.NewTimer(waitDuration)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// GetRate returns the current rate in bytes per second.
func (l *Limiter) GetRate() float64 {
	if l == nil {
		return 0
	}
	return l.rate
}
