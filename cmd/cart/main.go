// 仅作模板。实现 cart 服务时请重命名为 main.go。
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/cart"
)
// TODO:
// 1. 加载 MINIMALL_CART_* 配置。
// 2. 初始化结构化日志与 OpenTelemetry。
// 3. 组装 internal/cart 依赖。
// 4. 启动文档约定的 HTTP/gRPC 服务与 NATS 消费者。
// 5. 实现优雅退出。
func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_cart?parseTime=true"
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{"message": "db down", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "pong", "db": "ok"})
	})

	h := cart.NewHandler(db)
	auth := cart.AuthRequired()
	r.GET("/api/v1/cart", auth, h.ListByUser)
	r.POST("/api/v1/cart", auth, h.UpsertItem)
	r.DELETE("/api/v1/cart/:sku_id", auth, h.DeleteItem)


	log.Println("cart on :8083")
	if err := r.Run(":8083"); err != nil {
		log.Fatal(err)
	}
}
