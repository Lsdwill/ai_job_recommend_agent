package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowAndRefill(t *testing.T) {
	rl := NewRateLimiter(2, 1) // 容量2，每秒补充1

	// 注入可控时间
	now := time.Unix(0, 0).UnixNano()
	rl.now = func() int64 { return now }
	rl.lastRefill = now
	rl.tokens = 2

	if !rl.Allow() {
		t.Fatalf("expected first Allow() true")
	}
	if !rl.Allow() {
		t.Fatalf("expected second Allow() true")
	}
	if rl.Allow() {
		t.Fatalf("expected third Allow() false (no tokens)")
	}

	// 过 1 秒应补充 1 个令牌
	now += int64(time.Second)
	if !rl.Allow() {
		t.Fatalf("expected Allow() true after refill")
	}
	if rl.Allow() {
		t.Fatalf("expected Allow() false again (tokens should be 0)")
	}
}
