package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	httpapi "github.com/your-org/go-minimall/internal/gateway/adapter/http"
	"github.com/your-org/go-minimall/internal/gateway/config"
)

func Run() {
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	userURL := "http://127.0.0.1:8081"
	productURL := "http://127.0.0.1:8082"
	cartURL := "http://127.0.0.1:8083"
	paymentURL := "http://127.0.0.1:8085"
	searchURL := "http://127.0.0.1:8084"
	couponURL := "http://127.0.0.1:8086"
	notificationURL := "http://127.0.0.1:8087"
	inventoryURL := "http://127.0.0.1:8088"
	orderURL := "http://127.0.0.1:8089"
	toCart := httpapi.ReverseProxy(cartURL)
	toInventory := httpapi.ReverseProxy(inventoryURL)
	toOrder := httpapi.ReverseProxy(orderURL)
	toUser := httpapi.ReverseProxy(userURL)
	toProduct := httpapi.ReverseProxy(productURL)
	toPayment := httpapi.ReverseProxy(paymentURL)
	toCoupon := httpapi.ReverseProxy(couponURL)
	toSearch := httpapi.ReverseProxy(searchURL)
	toNotification := httpapi.ReverseProxy(notificationURL)

	auth := httpapi.AuthRequired()
	rateLimit := httpapi.RateLimitFromEnv()

	api := r.Group("/api/v1")
	api.Use(rateLimit)

	// 公开：auth
	api.POST("/auth/register", toUser)
	api.POST("/auth/login", toUser)
	api.POST("/auth/refresh", toUser)

	// 公开：商品读、搜索
	api.GET("/search/products", toSearch)
	api.GET("/categories", toProduct)
	api.GET("/products", toProduct)
	api.GET("/products/:id", toProduct)

	// 需登录：user
	api.GET("/users/me", auth, toUser)
	api.GET("/users/me/addresses", auth, toUser)
	api.POST("/users/me/addresses", auth, toUser)
	api.PUT("/users/me/addresses/:id", auth, toUser)
	api.DELETE("/users/me/addresses/:id", auth, toUser)

	// 需登录：购物车
	api.GET("/cart", auth, toCart)
	api.POST("/cart", auth, toCart)
	api.DELETE("/cart/:sku_id", auth, toCart)

	// 需登录：库存
	api.GET("/inventories/:sku_id", auth, toInventory)
	api.POST("/inventories", auth, toInventory)
	api.POST("/inventories/reserve", auth, toInventory)
	api.POST("/inventories/release", auth, toInventory)
	api.POST("/inventories/commit", auth, toInventory)

	// 需登录：订单
	api.POST("/orders", auth, toOrder)
	api.POST("/orders/:id/cancel", auth, toOrder)
	// 需登录：支付
	api.POST("/payments", auth, toPayment)
	api.POST("/payments/:id/mock-confirm", auth, toPayment)

	// 优惠券：模板公开读；领取、我的券、试算要登录。锁定/解锁/核销只给订单服务直连 :8086。
	api.GET("/coupons/templates", toCoupon)
	api.POST("/coupons/templates", auth, toCoupon)
	api.POST("/coupons/templates/:id/claim", auth, toCoupon)
	api.GET("/coupons/mine", auth, toCoupon)
	api.POST("/coupons/quote", auth, toCoupon)

	// 站内信。写入只给订单服务直连 :8087。
	api.GET("/notifications/mine", auth, toNotification)
	api.POST("/notifications/:id/read", auth, toNotification)
	// 需登录：商品写（学习阶段当 admin）
	api.POST("/categories", auth, toProduct)
	api.POST("/products", auth, toProduct)
	api.POST("/products/:id/skus", auth, toProduct)

	log.Println("gateway on " + config.HTTPAddr())
	if err := r.Run(config.HTTPAddr()); err != nil {
		log.Fatal(err)
	}
}
