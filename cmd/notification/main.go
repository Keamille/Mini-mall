package main

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/your-org/go-minimall/internal/notification"
)

func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_notification?parseTime=true"
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

	h := notification.NewHandler(db)
	auth := notification.AuthRequired()
	r.POST("/api/v1/notifications/events", h.Record)
	r.GET("/api/v1/notifications/mine", auth, h.ListMine)
	r.POST("/api/v1/notifications/:id/read", auth, h.MarkRead)

	log.Println("notification on :8087")
	if err := r.Run(":8087"); err != nil {
		log.Fatal(err)
	}
}
