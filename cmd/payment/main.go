package main

import (
	"log"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/payment"
)

func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_payment?parseTime=true"
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{"message": "db down", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "pong", "db": "ok"})
	})

	h := payment.NewHandler(db)
	auth := payment.AuthRequired()
	r.POST("/api/v1/payments", auth, h.CreatePayment)
	r.POST("/api/v1/payments/:id/mock-confirm", auth, h.MockConfirm)

	r.Run(":8085")
}
