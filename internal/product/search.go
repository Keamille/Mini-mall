package product

import (
	"context"
	"database/sql"
	"log"
	"time"

	searchv1 "github.com/your-org/go-minimall/api/gen/search/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (h *Handler) indexProduct(p productRow) {
	var minPrice sql.NullInt64
	_ = h.DB.Get(&minPrice, `
		SELECT MIN(price_cent) FROM skus WHERE product_id = ? AND status = 1`, p.ID)
	req := &searchv1.UpsertProductRequest{
		ProductId:    p.ID,
		Title:        p.Title,
		Subtitle:     p.Subtitle,
		Status:       int32(p.Status),
		MinPriceCent: minPrice.Int64,
	}
	if p.CategoryID.Valid {
		req.CategoryId = p.CategoryID.String
	}
	conn, err := grpc.NewClient("127.0.0.1:9007", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("search index %s: %v", p.ID, err)
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := searchv1.NewSearchServiceClient(conn).UpsertProduct(ctx, req); err != nil {
		log.Printf("search index %s: %v", p.ID, err)
	}
}
