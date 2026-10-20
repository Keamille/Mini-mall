package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitBlocksOverQuota(t *testing.T) {
	addr := getenv("MINIMALL_GATEWAY_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	rdb, err := NewRedisClient(addr)
	if err != nil {
		t.Skip("redis not available:", err)
	}
	defer rdb.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(rdb, RateLimitConfig{Limit: 2, Window: time.Minute}))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.RemoteAddr = "127.0.0.1:59999"
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: want 200 got %d", i+1, w.Code)
		}
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = "127.0.0.1:59999"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("third request: want 429 got %d", w.Code)
	}
}
