package middleware

import (
	"net/http"
	"qd-sc/internal/model"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 基于令牌桶的限流器（使用原子操作）
type RateLimiter struct {
	tokens     int64 // 当前令牌数
	capacity   int64 // 桶容量（最大突发请求数）
	refillRate int64 // 每秒补充的令牌数（持续QPS）
	lastRefill int64 // 上次补充时间（纳秒时间戳）

	now func() int64 // 便于测试注入（UnixNano）
}

// NewRateLimiter 创建限流器
// capacity: 桶容量（最大突发请求数）
// refillRate: 每秒补充的令牌数（持续QPS）
func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	return &RateLimiter{
		tokens:     int64(capacity),
		capacity:   int64(capacity),
		refillRate: int64(refillRate),
		lastRefill: time.Now().UnixNano(),
		now: func() int64 {
			return time.Now().UnixNano()
		},
	}
}

// Allow 尝试消耗一个令牌（使用CAS无锁算法）
func (rl *RateLimiter) Allow() bool {
	now := rl.now()

	for {
		// 读取当前状态
		currentTokens := atomic.LoadInt64(&rl.tokens)
		lastRefill := atomic.LoadInt64(&rl.lastRefill)

		// 计算应该补充的令牌
		elapsed := now - lastRefill
		if elapsed < 0 {
			// 时钟回拨等极端情况：不补充
			elapsed = 0
		}

		// 安全计算：避免 elapsed * refillRate 直接相乘造成溢出
		// tokensToAdd = floor(elapsed_ns * refillRate_per_sec / 1e9)
		secPart := elapsed / int64(time.Second)  // elapsed 秒
		nsecPart := elapsed % int64(time.Second) // 剩余纳秒
		tokensToAdd := secPart*rl.refillRate + (nsecPart*rl.refillRate)/int64(time.Second)

		newTokens := currentTokens
		if tokensToAdd > 0 {
			newTokens = currentTokens + tokensToAdd
			if newTokens > rl.capacity {
				newTokens = rl.capacity
			}
		}

		// 检查是否有令牌可用
		if newTokens < 1 {
			return false
		}

		// 尝试消耗一个令牌
		if atomic.CompareAndSwapInt64(&rl.tokens, currentTokens, newTokens-1) {
			// 更新最后补充时间
			if tokensToAdd > 0 {
				atomic.StoreInt64(&rl.lastRefill, now)
			}
			return true
		}

		// CAS失败，重试
	}
}

// RateLimit 限流中间件
func RateLimit(capacity, refillRate int) gin.HandlerFunc {
	limiter := NewRateLimiter(capacity, refillRate)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: "请求过于频繁，请稍后再试",
					Type:    "rate_limit_exceeded",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
