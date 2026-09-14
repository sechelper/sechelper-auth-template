package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLimiterReturns429DecisionAfterLimit(t *testing.T) {
	limiter := NewMemory()
	allowed, _, err := limiter.Allow(context.Background(), "login:test", 2, time.Minute)
	if err != nil || !allowed {
		t.Fatalf("first Allow() = (%v, %v), want allowed", allowed, err)
	}
	allowed, _, err = limiter.Allow(context.Background(), "login:test", 2, time.Minute)
	if err != nil || !allowed {
		t.Fatalf("second Allow() = (%v, %v), want allowed", allowed, err)
	}
	allowed, retryAfter, err := limiter.Allow(context.Background(), "login:test", 2, time.Minute)
	if err != nil || allowed || retryAfter <= 0 {
		t.Fatalf("third Allow() = (%v, %v, %v), want blocked with retry", allowed, retryAfter, err)
	}
}
