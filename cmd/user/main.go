package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/user"
)

// TODO:
// 1. 加载 MINIMALL_USER_* 配置。
// 2. 初始化结构化日志与 OpenTelemetry。
// 3. 组装 internal/user 依赖。
// 4. 启动文档约定的 HTTP/gRPC 服务与 NATS 消费者。
// 5. 实现优雅退出。
func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_user?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping MySQL: %v", err)
	}

	log.Println("Connected to MySQL")

	r := gin.Default()
	h := user.NewHandler(db)

	r.GET("/ping", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{"message": "db down", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "pong", "db": "ok"})
	})

	r.POST("/api/v1/auth/register", h.Register_user)
	r.POST("/api/v1/auth/login", h.Login)
	r.POST("/api/v1/auth/refresh", h.Refresh)
	auth := r.Group("/api/v1")
	auth.Use(h.AuthRequired())
	auth.GET("/users/me", h.Me)
	auth.GET("/users/me/addresses", h.ListAddress)
	auth.POST("/users/me/addresses", h.CreateAddress)
	auth.PUT("/users/me/addresses/:id", h.UpdateAddress)
	auth.DELETE("/users/me/addresses/:id", h.DeleteAddress)

	log.Print("Server started on port 8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
