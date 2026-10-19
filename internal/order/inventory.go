package order

import (
	"context"
	"time"

	inventoryv1 "github.com/your-org/go-minimall/api/gen/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func reserve(orderID string, items []itemRow) (string, int, error) {
	client, done, err := inventoryClient()
	if err != nil {
		return "", 0, err
	}
	defer done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &inventoryv1.ReserveStockRequest{OrderId: orderID}
	for _, it := range items {
		req.Items = append(req.Items, &inventoryv1.ReserveItem{
			SkuId:    it.SkuID,
			Quantity: int32(it.Quantity),
		})
	}
	resp, err := client.ReserveStock(ctx, req)
	if err != nil {
		if status.Code(err) == codes.FailedPrecondition {
			return "", 409, nil
		}
		return "", 0, err
	}
	return resp.GetReservationId(), 200, nil
}

func release(orderID string) error {
	client, done, err := inventoryClient()
	if err != nil {
		return err
	}
	defer done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.ReleaseStock(ctx, &inventoryv1.ReleaseStockRequest{OrderId: orderID})
	return err
}

func commitStock(orderID string) error {
	client, done, err := inventoryClient()
	if err != nil {
		return err
	}
	defer done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.CommitStock(ctx, &inventoryv1.CommitStockRequest{OrderId: orderID})
	return err
}

func inventoryClient() (inventoryv1.InventoryServiceClient, func(), error) {
	conn, err := grpc.NewClient("127.0.0.1:9008", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return inventoryv1.NewInventoryServiceClient(conn), func() { _ = conn.Close() }, nil
}
