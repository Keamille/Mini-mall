package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func lockCoupon(userID, userCouponID, orderID string, itemAmount int64) (int64, int, int, string, error) {
	body := map[string]any{
		"user_coupon_id":   userCouponID,
		"user_id":          userID,
		"order_id":         orderID,
		"item_amount_cent": itemAmount,
	}
	raw, _ := json.Marshal(body)
	resp, err := http.Post("http://127.0.0.1:8086/api/v1/coupons/lock", "application/json", bytes.NewReader(raw))
	if err != nil {
		return 0, 0, 0, "", err
	}
	defer resp.Body.Close()

	var out struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			DiscountCent int64 `json:"discount_cent"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, resp.StatusCode, 0, "", err
	}
	return out.Data.DiscountCent, resp.StatusCode, out.Code, out.Message, nil
}

func unlockCoupon(userID, userCouponID, orderID string) error {
	return postCoupon("http://127.0.0.1:8086/api/v1/coupons/unlock", userID, userCouponID, orderID)
}

func confirmCoupon(userID, userCouponID, orderID string) error {
	return postCoupon("http://127.0.0.1:8086/api/v1/coupons/confirm", userID, userCouponID, orderID)
}

func postCoupon(url, userID, userCouponID, orderID string) error {
	body := map[string]any{
		"user_coupon_id": userCouponID,
		"user_id":        userID,
		"order_id":       orderID,
	}
	raw, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coupon %s: %d", url, resp.StatusCode)
	}
	return nil
}
