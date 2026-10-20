package main

import (
	"log"
	"net"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	notificationv1 "github.com/your-org/go-minimall/api/gen/notification/v1"
	"github.com/your-org/go-minimall/internal/notification"
	"google.golang.org/grpc"
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

	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Printf("notification nats disabled: %v", err)
	} else {
		js, err := jetstream.New(nc)
		if err != nil {
			log.Printf("notification jetstream disabled: %v", err)
		} else if err := notification.StartOrderCreatedConsumer(h, js); err != nil {
			log.Printf("notification nats consumer disabled: %v", err)
		} else {
			log.Println("notification nats consumer on ORDERS/notify-worker")
		}
	}

	auth := notification.AuthRequired()
	r.POST("/api/v1/notifications/events", h.Record)
	r.GET("/api/v1/notifications/mine", auth, h.ListMine)
	r.POST("/api/v1/notifications/:id/read", auth, h.MarkRead)

	lis, err := net.Listen("tcp", ":9009")
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer()
	notificationv1.RegisterNotificationServiceServer(gs, notification.NewGRPCServer(h))
	go func() {
		log.Println("notification grpc on :9009")
		if err := gs.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("notification on :8087")
	if err := r.Run(":8087"); err != nil {
		log.Fatal(err)
	}
}
