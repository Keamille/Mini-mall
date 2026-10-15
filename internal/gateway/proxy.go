package gateway

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func ReverseProxy(target string) gin.HandlerFunc {
	u, _ := url.Parse(target) // 学习阶段可先忽略 error
	proxy := httputil.NewSingleHostReverseProxy(u)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
