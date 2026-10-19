package main

import (
	"log"
	"net"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	searchv1 "github.com/your-org/go-minimall/api/gen/search/v1"
	"github.com/your-org/go-minimall/internal/search"
	"google.golang.org/grpc"
)

func main() {
	dsn := "minimall:123456@tcp(127.0.0.1:3306)/minimall_search?parseTime=true"
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

	h := search.NewHandler(db)
	r.GET("/api/v1/search/products", h.SearchProducts)
	r.POST("/api/v1/search/products", h.Upsert)

	lis, err := net.Listen("tcp", ":9007")
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer()
	searchv1.RegisterSearchServiceServer(gs, search.NewGRPCServer(h))
	go func() {
		log.Println("search grpc on :9007")
		if err := gs.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("search on :8084")
	if err := r.Run(":8084"); err != nil {
		log.Fatal(err)
	}
}
