package inventory

import (
	"context"
	"errors"
	"time"

	inventoryv1 "github.com/your-org/go-minimall/api/gen/inventory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	inventoryv1.UnimplementedInventoryServiceServer
	h *Handler
}

func NewGRPCServer(h *Handler) inventoryv1.InventoryServiceServer {
	return &grpcServer{h: h}
}

func (s *grpcServer) ReserveStock(ctx context.Context, req *inventoryv1.ReserveStockRequest) (*inventoryv1.ReserveStockResponse, error) {
	if req.GetOrderId() == "" || len(req.GetItems()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	items := make([]reserveItem, 0, len(req.GetItems()))
	for _, it := range req.GetItems() {
		if it.GetQuantity() <= 0 {
			return nil, status.Error(codes.InvalidArgument, "quantity must be > 0")
		}
		items = append(items, reserveItem{SkuID: it.GetSkuId(), Quantity: int(it.GetQuantity())})
	}

	id, err := s.h.reserveTx(req.GetOrderId(), time.Now().Add(15*time.Minute), items)
	if errors.Is(err, ErrInsufficientStock) {
		return nil, status.Error(codes.FailedPrecondition, "insufficient stock")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to reserve")
	}
	return &inventoryv1.ReserveStockResponse{
		ReservationId: id,
		OrderId:       req.GetOrderId(),
	}, nil
}

func (s *grpcServer) ReleaseStock(ctx context.Context, req *inventoryv1.ReleaseStockRequest) (*inventoryv1.ReleaseStockResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	err := s.h.releaseTx(req.GetOrderId())
	if errors.Is(err, ErrReservationNotFound) {
		return nil, status.Error(codes.NotFound, "reservation not found")
	}
	if errors.Is(err, ErrReservationCommitted) {
		return nil, status.Error(codes.FailedPrecondition, "reservation already committed")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to release")
	}
	return &inventoryv1.ReleaseStockResponse{}, nil
}

func (s *grpcServer) CommitStock(ctx context.Context, req *inventoryv1.CommitStockRequest) (*inventoryv1.CommitStockResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	err := s.h.commitTx(req.GetOrderId())
	if errors.Is(err, ErrReservationNotFound) {
		return nil, status.Error(codes.NotFound, "reservation not found")
	}
	if errors.Is(err, ErrReservationCommitted) {
		return nil, status.Error(codes.FailedPrecondition, "reservation already committed")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to commit")
	}
	return &inventoryv1.CommitStockResponse{}, nil
}
