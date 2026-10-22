package grpcadp

import (
	"context"
	"errors"
	"time"

	orderv1 "github.com/your-org/go-minimall/api/gen/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NotifyOrderPaid(orderID string) error {
	conn, err := grpc.NewClient("127.0.0.1:9004", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = orderv1.NewOrderServiceClient(conn).MarkPaid(ctx, &orderv1.MarkPaidRequest{OrderId: orderID})
	if err != nil {
		return errors.New("order paid failed")
	}
	return nil
}
