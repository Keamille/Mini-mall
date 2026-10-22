package server

import (
	"log"

	"context"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	orderv1 "github.com/your-org/go-minimall/api/gen/order/v1"
	grpcadp "github.com/your-org/go-minimall/internal/order/adapter/grpc"
	httpapi "github.com/your-org/go-minimall/internal/order/adapter/http"
	mysqlrepo "github.com/your-org/go-minimall/internal/order/adapter/mysql"
	natsadp "github.com/your-org/go-minimall/internal/order/adapter/nats"
	"github.com/your-org/go-minimall/internal/order/app"
	"github.com/your-org/go-minimall/internal/order/config"
	"google.golang.org/grpc"
	"net"
	"time"
)

func Run() {
	cfg := config.Load()
	db, err := sqlx.Connect("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()
	mysqlrepo.EnsureSchema(db)
	repo := mysqlrepo.NewRepo(db)
	var js jetstream.JetStream
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Printf("order nats disabled: %v", err)
	} else {
		j, err := jetstream.New(nc)
		if err != nil {
			log.Printf("order jetstream disabled: %v", err)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if _, err := j.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
				Name:     "ORDERS",
				Subjects: []string{"minimall.order.>"},
			}); err != nil {
				log.Printf("order stream ORDERS: %v", err)
			} else {
				js = j
				log.Println("order nats stream ORDERS ready")
			}
			cancel()
		}
	}
	h := app.NewHandler(repo, grpcadp.InventoryClient{}, grpcadp.CouponClient{}, grpcadp.Notifier{}, natsadp.Publisher{JS: js})
	h.StartCloseWorker()
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
	orderv1.RegisterOrderServiceServer(gs, grpcadp.NewGRPCServer(h))
	go func() {
		log.Println("order grpc on " + cfg.GRPCAddr)
		if err := gs.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("order on " + cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
