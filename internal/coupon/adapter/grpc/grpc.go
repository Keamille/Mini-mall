package grpcadp

import (
	"context"
	"errors"
	"fmt"
	"github.com/your-org/go-minimall/internal/coupon/app"
	"github.com/your-org/go-minimall/internal/coupon/domain"

	couponv1 "github.com/your-org/go-minimall/api/gen/coupon/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	couponv1.UnimplementedCouponServiceServer
	h *app.Handler
}

func NewGRPCServer(h *app.Handler) couponv1.CouponServiceServer {
	return &grpcServer{h: h}
}

func (s *grpcServer) LockCoupon(ctx context.Context, req *couponv1.LockCouponRequest) (*couponv1.LockCouponResponse, error) {
	in := app.AmountRequest{
		UserCouponID:   req.GetUserCouponId(),
		UserID:         req.GetUserId(),
		OrderID:        req.GetOrderId(),
		ItemAmountCent: req.GetItemAmountCent(),
	}
	if !app.ValidLockRequest(in) {
		return nil, status.Error(codes.InvalidArgument, "10001 invalid request")
	}
	out, err := s.h.Repo.LockTx(in.UserID, in.UserCouponID, in.OrderID, in.ItemAmountCent)
	if err != nil {
		return nil, couponStatus(err)
	}
	return &couponv1.LockCouponResponse{DiscountCent: out.DiscountCent}, nil
}

func (s *grpcServer) UnlockCoupon(ctx context.Context, req *couponv1.CouponOrderRequest) (*couponv1.CouponOrderResponse, error) {
	if !validCouponOrder(req) {
		return nil, status.Error(codes.InvalidArgument, "10001 invalid request")
	}
	if err := s.h.Repo.UnlockTx(req.GetUserId(), req.GetUserCouponId(), req.GetOrderId()); err != nil {
		return nil, couponStatus(err)
	}
	return &couponv1.CouponOrderResponse{}, nil
}

func (s *grpcServer) ConfirmCoupon(ctx context.Context, req *couponv1.CouponOrderRequest) (*couponv1.CouponOrderResponse, error) {
	if !validCouponOrder(req) {
		return nil, status.Error(codes.InvalidArgument, "10001 invalid request")
	}
	if err := s.h.Repo.ConfirmTx(req.GetUserId(), req.GetUserCouponId(), req.GetOrderId()); err != nil {
		return nil, couponStatus(err)
	}
	return &couponv1.CouponOrderResponse{}, nil
}

func validCouponOrder(req *couponv1.CouponOrderRequest) bool {
	return req.GetUserCouponId() != "" && req.GetUserId() != "" && len(req.GetOrderId()) == 26
}

func couponStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrTemplateNotClaimable):
		return status.Error(codes.NotFound, "17001 template not claimable")
	case errors.Is(err, domain.ErrClaimLimit):
		return status.Error(codes.FailedPrecondition, "17002 claim limit reached")
	case errors.Is(err, domain.ErrCouponNotFound):
		return status.Error(codes.NotFound, "17003 coupon not found")
	case errors.Is(err, domain.ErrCouponStatus):
		return status.Error(codes.FailedPrecondition, "17004 coupon status invalid")
	case errors.Is(err, domain.ErrCouponThreshold):
		return status.Error(codes.InvalidArgument, "17005 threshold not met")
	case errors.Is(err, domain.ErrCouponExpired):
		return status.Error(codes.InvalidArgument, "17006 coupon expired")
	default:
		return status.Error(codes.Internal, fmt.Sprintf("10007 coupon request failed: %v", err))
	}
}
