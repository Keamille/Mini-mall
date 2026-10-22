package server

import (
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	httpapi "github.com/your-org/go-minimall/internal/user/adapter/http"
	mysqlrepo "github.com/your-org/go-minimall/internal/user/adapter/mysql"
	"github.com/your-org/go-minimall/internal/user/app"
	"github.com/your-org/go-minimall/internal/user/config"
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
	log.Println("Connected to MySQL")
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{"message": "db down", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "pong", "db": "ok"})
	})
	httpapi.Register(r, h)
	log.Println("user on " + cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
