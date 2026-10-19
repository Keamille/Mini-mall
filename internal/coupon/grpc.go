package coupon

import (
	"context"
	"errors"
	"fmt"

	couponv1 "github.com/your-org/go-minimall/api/gen/coupon/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	couponv1.UnimplementedCouponServiceServer
	h *Handler
}

func NewGRPCServer(h *Handler) couponv1.CouponServiceServer {
	return &grpcServer{h: h}
}

func (s *grpcServer) LockCoupon(ctx context.Context, req *couponv1.LockCouponRequest) (*couponv1.LockCouponResponse, error) {
	in := amountRequest{
		UserCouponID:   req.GetUserCouponId(),
		UserID:         req.GetUserId(),
		OrderID:        req.GetOrderId(),
		ItemAmountCent: req.GetItemAmountCent(),
	}
	if !validLockRequest(in) {
		return nil, status.Error(codes.InvalidArgument, "10001 invalid request")
	}
	out, err := s.h.lockTx(in.UserID, in.UserCouponID, in.OrderID, in.ItemAmountCent)
	if err != nil {
		return nil, couponStatus(err)
	}
	return &couponv1.LockCouponResponse{DiscountCent: out.DiscountCent}, nil
}

func (s *grpcServer) UnlockCoupon(ctx context.Context, req *couponv1.CouponOrderRequest) (*couponv1.CouponOrderResponse, error) {
	if !validCouponOrder(req) {
		return nil, status.Error(codes.InvalidArgument, "10001 invalid request")
	}
	if err := s.h.unlockTx(req.GetUserId(), req.GetUserCouponId(), req.GetOrderId()); err != nil {
		return nil, couponStatus(err)
	}
	return &couponv1.CouponOrderResponse{}, nil
}

func (s *grpcServer) ConfirmCoupon(ctx context.Context, req *couponv1.CouponOrderRequest) (*couponv1.CouponOrderResponse, error) {
	if !validCouponOrder(req) {
		return nil, status.Error(codes.InvalidArgument, "10001 invalid request")
	}
	if err := s.h.confirmTx(req.GetUserId(), req.GetUserCouponId(), req.GetOrderId()); err != nil {
		return nil, couponStatus(err)
	}
	return &couponv1.CouponOrderResponse{}, nil
}

func validCouponOrder(req *couponv1.CouponOrderRequest) bool {
	return req.GetUserCouponId() != "" && req.GetUserId() != "" && len(req.GetOrderId()) == 26
}

func couponStatus(err error) error {
	switch {
	case errors.Is(err, ErrTemplateNotClaimable):
		return status.Error(codes.NotFound, "17001 template not claimable")
	case errors.Is(err, ErrClaimLimit):
		return status.Error(codes.FailedPrecondition, "17002 claim limit reached")
	case errors.Is(err, ErrCouponNotFound):
		return status.Error(codes.NotFound, "17003 coupon not found")
	case errors.Is(err, ErrCouponStatus):
		return status.Error(codes.FailedPrecondition, "17004 coupon status invalid")
	case errors.Is(err, ErrCouponThreshold):
		return status.Error(codes.InvalidArgument, "17005 threshold not met")
	case errors.Is(err, ErrCouponExpired):
		return status.Error(codes.InvalidArgument, "17006 coupon expired")
	default:
		return status.Error(codes.Internal, fmt.Sprintf("10007 coupon request failed: %v", err))
	}
}
