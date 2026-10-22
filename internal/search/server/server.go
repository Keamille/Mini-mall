package server

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/go-minimall/internal/search/adapter/grpc"
	httpapi "github.com/your-org/go-minimall/internal/search/adapter/http"
	"github.com/your-org/go-minimall/internal/search/adapter/opensearch"
	"github.com/your-org/go-minimall/internal/search/app"
	"github.com/your-org/go-minimall/internal/search/config"
	"google.golang.org/grpc"
)

func Run() {
	cfgURL := config.OpenSearchURL()
	client := opensearch.New(cfgURL, config.IndexAlias())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("opensearch %s: %v", cfgURL, err)
	}
	if err := client.Ensure(ctx); err != nil {
		log.Fatalf("opensearch index: %v", err)
	}
	log.Printf("search index alias %s on %s", config.IndexAlias(), cfgURL)

	svc := app.NewService(client)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		pctx, pcancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer pcancel()
		if err := client.Ping(pctx); err != nil {
			c.JSON(500, gin.H{"message": "opensearch down", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "pong", "opensearch": "ok"})
	})
	httpapi.Register(r, svc)

	lis, err := net.Listen("tcp", config.GRPCAddr())
	if err != nil {
		log.Fatal(err)
	}
	gs := grpc.NewServer()
	grpcadp.Register(gs, svc)
	go func() {
		log.Println("search grpc on " + config.GRPCAddr())
		if err := gs.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("search on " + config.HTTPAddr())
	if err := r.Run(config.HTTPAddr()); err != nil {
		log.Fatal(err)
	}
}
