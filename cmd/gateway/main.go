package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/gateway"
)

func main() {
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	userURL := "http://127.0.0.1:8081"
	productURL := "http://127.0.0.1:8082"
	cartURL := "http://127.0.0.1:8083"
	inventoryURL := "http://127.0.0.1:8088"
	orderURL := "http://127.0.0.1:8089"
	toCart := gateway.ReverseProxy(cartURL)
	toInventory := gateway.ReverseProxy(inventoryURL)
	toOrder := gateway.ReverseProxy(orderURL)
	toUser := gateway.ReverseProxy(userURL)
	toProduct := gateway.ReverseProxy(productURL)
	auth := gateway.AuthRequired()

	api := r.Group("/api/v1")

	// 公开：auth
	api.POST("/auth/register", toUser)
	api.POST("/auth/login", toUser)
	api.POST("/auth/refresh", toUser)

	// 公开：商品读
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

	// 需登录：订单
	api.POST("/orders", auth, toOrder)
	api.POST("/orders/:id/cancel", auth, toOrder)

	// 需登录：商品写（学习阶段当 admin）
	api.POST("/categories", auth, toProduct)
	api.POST("/products", auth, toProduct)
	api.POST("/products/:id/skus", auth, toProduct)

	log.Println("gateway on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
