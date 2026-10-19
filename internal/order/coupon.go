package order

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	couponv1 "github.com/your-org/go-minimall/api/gen/coupon/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func lockCoupon(userID, userCouponID, orderID string, itemAmount int64) (int64, int, int, string, error) {
	client, done, err := couponClient()
	if err != nil {
		return 0, 0, 0, "", err
	}
	defer done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.LockCoupon(ctx, &couponv1.LockCouponRequest{
		UserCouponId:   userCouponID,
		UserId:         userID,
		OrderId:        orderID,
		ItemAmountCent: itemAmount,
	})
	if err != nil {
		statusCode, code, message, ok := couponHTTP(err)
		if !ok {
			return 0, 0, 0, "", err
		}
		return 0, statusCode, code, message, nil
	}
	return resp.GetDiscountCent(), http.StatusOK, 0, "", nil
}

func unlockCoupon(userID, userCouponID, orderID string) error {
	return couponOrder(func(ctx context.Context, client couponv1.CouponServiceClient) error {
		_, err := client.UnlockCoupon(ctx, couponOrderReq(userID, userCouponID, orderID))
		return err
	})
}

func confirmCoupon(userID, userCouponID, orderID string) error {
	return couponOrder(func(ctx context.Context, client couponv1.CouponServiceClient) error {
		_, err := client.ConfirmCoupon(ctx, couponOrderReq(userID, userCouponID, orderID))
		return err
	})
}

func couponOrder(fn func(context.Context, couponv1.CouponServiceClient) error) error {
	client, done, err := couponClient()
	if err != nil {
		return err
	}
	defer done()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return fn(ctx, client)
}

func couponOrderReq(userID, userCouponID, orderID string) *couponv1.CouponOrderRequest {
	return &couponv1.CouponOrderRequest{UserId: userID, UserCouponId: userCouponID, OrderId: orderID}
}

func couponClient() (couponv1.CouponServiceClient, func(), error) {
	conn, err := grpc.NewClient("127.0.0.1:9006", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return couponv1.NewCouponServiceClient(conn), func() { _ = conn.Close() }, nil
}

func couponHTTP(err error) (statusCode, code int, message string, ok bool) {
	st, ok := status.FromError(err)
	if !ok {
		return 0, 0, "", false
	}
	head, msg, _ := strings.Cut(st.Message(), " ")
	n, convErr := strconv.Atoi(head)
	if convErr != nil {
		return 0, 0, "", false
	}
	switch n {
	case 17001, 17003:
		statusCode = http.StatusNotFound
	case 17002, 17004:
		statusCode = http.StatusConflict
	case 10001, 17005, 17006:
		statusCode = http.StatusBadRequest
	default:
		statusCode = http.StatusInternalServerError
	}
	return statusCode, n, msg, true
}
