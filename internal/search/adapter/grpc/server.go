package grpcadp

import (
	"context"
	"errors"

	searchv1 "github.com/your-org/go-minimall/api/gen/search/v1"
	"github.com/your-org/go-minimall/internal/search/app"
	"github.com/your-org/go-minimall/internal/search/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	searchv1.UnimplementedSearchServiceServer
	svc *app.Service
}

func Register(gs *grpc.Server, svc *app.Service) {
	searchv1.RegisterSearchServiceServer(gs, &server{svc: svc})
}

func (s *server) UpsertProduct(ctx context.Context, req *searchv1.UpsertProductRequest) (*searchv1.UpsertProductResponse, error) {
	doc := domain.Document{
		ProductID:    req.GetProductId(),
		CategoryID:   req.GetCategoryId(),
		Title:        req.GetTitle(),
		Subtitle:     req.GetSubtitle(),
		Status:       int(req.GetStatus()),
		MinPriceCent: req.GetMinPriceCent(),
	}
	if err := s.svc.Upsert(ctx, doc); err != nil {
		if errors.Is(err, domain.ErrBadDocument) {
			return nil, status.Error(codes.InvalidArgument, "invalid search document")
		}
		return nil, status.Error(codes.Internal, "search unavailable")
	}
	return &searchv1.UpsertProductResponse{}, nil
}
