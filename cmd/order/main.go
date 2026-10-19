package main

import (
	"log"
	"net"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	orderv1 "github.com/your-org/go-minimall/api/gen/order/v1"
	"github.com/your-org/go-minimall/internal/order"
	"google.golang.org/grpc"
)

func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_order?parseTime=true"

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
	auth := order.AuthRequired()
	h := order.NewHandler(db)

	r.POST("/api/v1/orders", auth, h.CreateOrder)
	r.POST("/api/v1/orders/:id/cancel", auth, h.CancelOrder)
	r.POST("/api/v1/orders/:id/paid", h.MarkPaid)

	lis, err := net.Listen("tcp", ":9004")
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(gs, order.NewGRPCServer(h))
	go func() {
		log.Println("order grpc on :9004")
		if err := gs.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("order on :8089")
	if err := r.Run(":8089"); err != nil {
		log.Fatal(err)
	}
}
