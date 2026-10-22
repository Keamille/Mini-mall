package server

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	inventoryv1 "github.com/your-org/go-minimall/api/gen/inventory/v1"
	grpcadp "github.com/your-org/go-minimall/internal/inventory/adapter/grpc"
	httpapi "github.com/your-org/go-minimall/internal/inventory/adapter/http"
	mysqlrepo "github.com/your-org/go-minimall/internal/inventory/adapter/mysql"
	"github.com/your-org/go-minimall/internal/inventory/app"
	"github.com/your-org/go-minimall/internal/inventory/config"
	"google.golang.org/grpc"
	"net"
)

func Run() {
	cfg := config.Load()
	db, err := sqlx.Connect("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()
	repo := mysqlrepo.NewRepo(db)
	h := app.NewHandler(repo)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{"message": "db down", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "pong", "db": "ok"})
	})
	httpapi.Register(r, h)
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer()
	inventoryv1.RegisterInventoryServiceServer(gs, grpcadp.NewGRPCServer(h))
	go func() {
		log.Println("inventory grpc on " + cfg.GRPCAddr)
		if err := gs.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("inventory on " + cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
