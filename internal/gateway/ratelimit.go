package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds per-window request quota for gateway clients.
type RateLimitConfig struct {
	Limit  int
	Window time.Duration
}

// NewRedisClient dials Redis for gateway middleware. Addr is host:port.
func NewRedisClient(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return rdb, nil
}

// RateLimit returns middleware that counts requests per client IP in Redis.
// Redis errors fail open so the gateway stays usable when Redis is down.
func RateLimit(rdb *redis.Client, cfg RateLimitConfig) gin.HandlerFunc {
	if cfg.Limit <= 0 {
		cfg.Limit = 120
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	return func(c *gin.Context) {
		key := fmt.Sprintf("gateway:rl:ip:%s", c.ClientIP())
		ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
		defer cancel()

		n, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("gateway ratelimit redis: %v", err)
			c.Next()
			return
		}
		if n == 1 {
			if err := rdb.Expire(ctx, key, cfg.Window).Err(); err != nil {
				log.Printf("gateway ratelimit expire: %v", err)
			}
		}
		if n > int64(cfg.Limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    10006,
				"message": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// RateLimitFromEnv builds middleware from MINIMALL_GATEWAY_REDIS_ADDR and optional
// MINIMALL_GATEWAY_RATE_LIMIT / MINIMALL_GATEWAY_RATE_WINDOW_SEC.
// Returns nil handler if addr is empty or Redis is unreachable (logged).
func RateLimitFromEnv() gin.HandlerFunc {
	addr := envOr("MINIMALL_GATEWAY_REDIS_ADDR", "127.0.0.1:6379")
	if addr == "" || addr == "off" {
		log.Println("gateway ratelimit disabled (MINIMALL_GATEWAY_REDIS_ADDR empty or off)")
		return func(c *gin.Context) { c.Next() }
	}
	limit, _ := strconv.Atoi(envOr("MINIMALL_GATEWAY_RATE_LIMIT", "120"))
	sec, _ := strconv.Atoi(envOr("MINIMALL_GATEWAY_RATE_WINDOW_SEC", "60"))
	rdb, err := NewRedisClient(addr)
	if err != nil {
		log.Printf("gateway ratelimit disabled: redis %s: %v", addr, err)
		return func(c *gin.Context) { c.Next() }
	}
	log.Printf("gateway ratelimit: redis=%s limit=%d window=%ds", addr, limit, sec)
	return RateLimit(rdb, RateLimitConfig{Limit: limit, Window: time.Duration(sec) * time.Second})
}

func envOr(key, def string) string {
	if v := getenv(key); v != "" {
		return v
	}
	return def
}

// getenv is overridden in tests; default uses os.Getenv.
var getenv = func(key string) string {
	return getEnv(key)
}
