package search

import (
	"context"

	searchv1 "github.com/your-org/go-minimall/api/gen/search/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	searchv1.UnimplementedSearchServiceServer
	h *Handler
}

func NewGRPCServer(h *Handler) searchv1.SearchServiceServer {
	return &grpcServer{h: h}
}

func (s *grpcServer) UpsertProduct(ctx context.Context, req *searchv1.UpsertProductRequest) (*searchv1.UpsertProductResponse, error) {
	doc := document{
		ProductID:    req.GetProductId(),
		CategoryID:   req.GetCategoryId(),
		Title:        req.GetTitle(),
		Subtitle:     req.GetSubtitle(),
		Status:       int(req.GetStatus()),
		MinPriceCent: req.GetMinPriceCent(),
	}
	if len(doc.ProductID) != 26 || doc.Title == "" || doc.Status < 0 || doc.Status > 2 || doc.MinPriceCent < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid search document")
	}
	if err := s.h.upsertDocument(doc); err != nil {
		return nil, status.Error(codes.Internal, "search unavailable")
	}
	return &searchv1.UpsertProductResponse{}, nil
}
