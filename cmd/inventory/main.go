// 仅作模板。实现 inventory 服务时请重命名为 main.go。
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/inventory"
)

// TODO:
// 1. 加载 MINIMALL_INVENTORY_* 配置。
// 2. 初始化结构化日志与 OpenTelemetry。
// 3. 组装 internal/inventory 依赖。
// 4. 启动文档约定的 HTTP/gRPC 服务与 NATS 消费者。
// 5. 实现优雅退出。
func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_inventory?parseTime=true"

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

	h := inventory.NewHandler(db)
	r.POST("/api/v1/inventories", h.InitStock)
	r.GET("/api/v1/inventories/:sku_id", h.GetStock)
	r.POST("/api/v1/inventories/reserve", h.Reserve)
	r.POST("/api/v1/inventories/release", h.Release)
	log.Println("inventory on :8088")
	if err := r.Run(":8088"); err != nil {
		log.Fatal(err)
	}

}
