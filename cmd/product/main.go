// Template only. Rename to main.go when implementing the product service.
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/product"
)

// TODO:
// 1. Load MINIMALL_PRODUCT_* configuration.
// 2. Initialize structured logging and OpenTelemetry.
// 3. Assemble internal/product dependencies.
// 4. Start the documented HTTP/gRPC server and NATS consumers.
// 5. Implement graceful shutdown.
func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_product?parseTime=true"
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
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
	h := product.NewHandler(db)

	r.GET("/api/v1/categories", h.ListCategories)
	r.POST("/api/v1/categories", h.CreateCategory)
	r.GET("/api/v1/products", h.ListProducts)
	r.POST("/api/v1/products", h.CreateProduct)
	r.GET("/api/v1/products/:id", h.GetProduct)
	r.POST("/api/v1/products/:id/skus", h.CreateSku)
	log.Printf("server is starting on port %s", ":8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
