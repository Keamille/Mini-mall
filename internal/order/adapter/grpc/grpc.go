package grpcadp

import (
	"context"
	"errors"
	"github.com/your-org/go-minimall/internal/order/app"

	orderv1 "github.com/your-org/go-minimall/api/gen/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	orderv1.UnimplementedOrderServiceServer
	h *app.Handler
}

func NewGRPCServer(h *app.Handler) orderv1.OrderServiceServer {
	return &grpcServer{h: h}
}

func (s *grpcServer) MarkPaid(ctx context.Context, req *orderv1.MarkPaidRequest) (*orderv1.MarkPaidResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if err := s.h.ApplyPaid(req.GetOrderId()); err != nil {
		if errors.Is(err, app.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		if errors.Is(err, app.ErrOrderCannotPay) {
			return nil, status.Error(codes.FailedPrecondition, "order cannot be paid")
		}
		return nil, status.Error(codes.Internal, "failed to mark paid")
	}
	return &orderv1.MarkPaidResponse{OrderId: req.GetOrderId(), Status: 20}, nil
}
