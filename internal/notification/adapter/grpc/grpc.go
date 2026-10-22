package grpcadp

import (
	"context"
	"errors"
	"github.com/your-org/go-minimall/internal/notification/app"
	"github.com/your-org/go-minimall/internal/notification/domain"

	notificationv1 "github.com/your-org/go-minimall/api/gen/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	notificationv1.UnimplementedNotificationServiceServer
	h *app.Handler
}

func NewGRPCServer(h *app.Handler) notificationv1.NotificationServiceServer {
	return &grpcServer{h: h}
}

func (s *grpcServer) RecordEvent(ctx context.Context, req *notificationv1.RecordEventRequest) (*notificationv1.RecordEventResponse, error) {
	if req.GetEventId() == "" || len(req.GetUserId()) != 26 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	payload := []byte(req.GetPayloadJson())
	if _, err := s.h.Repo.Record(req.GetEventId(), req.GetEventType(), req.GetUserId(), payload); err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return nil, status.Error(codes.InvalidArgument, "template not found")
		}
		return nil, status.Error(codes.Internal, "failed to send")
	}
	return &notificationv1.RecordEventResponse{}, nil
}
