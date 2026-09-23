package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const fixedWindowScript = `
local count = redis.call("INCR", KEYS[1])
if count == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count`

type RateLimiter struct {
	rdb    redis.Cmdable
	limit  int64
	window time.Duration
	script *redis.Script
}

func NewRateLimiter(rdb redis.Cmdable, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		limit:  limit,
		window: window,
		script: redis.NewScript(fixedWindowScript),
	}
}

func (rl *RateLimiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ratelimit:" + c.ClientIP()

		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Millisecond)
		defer cancel()

		count, err := rl.script.Run(ctx, rl.rdb, []string{key}, rl.window.Milliseconds()).Int64()
		if err != nil {
			slog.Warn("rate limiter unavailable, allowing request", "err", err)
			c.Next()
			return
		}

		remaining := rl.limit - count
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.FormatInt(rl.limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

		if count > rl.limit {
			c.Header("Retry-After", strconv.FormatInt(int64(rl.window.Seconds()), 10))
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}