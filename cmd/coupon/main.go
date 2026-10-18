package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/coupon"
)

func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_coupon?parseTime=true"
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
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

	h := coupon.NewHandler(db)
	auth := coupon.AuthRequired()
	r.GET("/api/v1/coupons/templates", h.ListTemplates)
	r.POST("/api/v1/coupons/templates", auth, h.CreateTemplate)
	r.POST("/api/v1/coupons/templates/:id/claim", auth, h.Claim)
	r.GET("/api/v1/coupons/mine", auth, h.ListMine)
	r.POST("/api/v1/coupons/quote", auth, h.Quote)
	r.POST("/api/v1/coupons/lock", h.Lock)
	r.POST("/api/v1/coupons/unlock", h.Unlock)
	r.POST("/api/v1/coupons/confirm", h.Confirm)

	log.Println("coupon on :8086")
	if err := r.Run(":8086"); err != nil {
		log.Fatal(err)
	}
}
