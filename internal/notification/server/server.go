package server

import (
	"log"
	"net"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	notificationv1 "github.com/your-org/go-minimall/api/gen/notification/v1"
	grpcadp "github.com/your-org/go-minimall/internal/notification/adapter/grpc"
	httpapi "github.com/your-org/go-minimall/internal/notification/adapter/http"
	mysqlrepo "github.com/your-org/go-minimall/internal/notification/adapter/mysql"
	natsadp "github.com/your-org/go-minimall/internal/notification/adapter/nats"
	"github.com/your-org/go-minimall/internal/notification/app"
	"github.com/your-org/go-minimall/internal/notification/config"
	"google.golang.org/grpc"
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
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Printf("notification nats disabled: %v", err)
	} else {
		js, err := jetstream.New(nc)
		if err != nil {
			log.Printf("notification jetstream disabled: %v", err)
		} else if err := natsadp.StartOrderCreatedConsumer(h, js); err != nil {
			log.Printf("notification nats consumer disabled: %v", err)
		} else {
			log.Println("notification nats consumer on ORDERS/notify-worker")
		}
	}
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
	notificationv1.RegisterNotificationServiceServer(gs, grpcadp.NewGRPCServer(h))
	go func() {
		log.Println("notification grpc on " + cfg.GRPCAddr)
		if err := gs.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("notification on " + cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
