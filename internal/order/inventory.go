package order

import (
	"bytes"
	"encoding/json"

	"net/http"
)

func reserve(orderID string, items []itemRow) (string, int, error) {
	body := map[string]any{"order_id": orderID, "items": []map[string]any{}}
	list := make([]map[string]any, 0, len(items))
	for _, it := range items {
		list = append(list, map[string]any{"sku_id": it.SkuID, "quantity": it.Quantity})
	}
	body["items"] = list
	raw, _ := json.Marshal(body)

	resp, err := http.Post("http://127.0.0.1:8088/api/v1/inventories/reserve", "application/json", bytes.NewReader(raw))
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	var out struct {
		Code int `json:"code"`
		Data struct {
			ReservationID string `json:"reservation_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", resp.StatusCode, err
	}
	return out.Data.ReservationID, resp.StatusCode, nil
}

func release(orderID string) error {
	body := map[string]any{"order_id": orderID}
	raw, _ := json.Marshal(body)
	resp, err := http.Post("http://127.0.0.1:8088/api/v1/inventories/release", "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func commitStock(orderID string) error {
	body := map[string]any{"order_id": orderID}
	raw, _ := json.Marshal(body)
	resp, err := http.Post("http://127.0.0.1:8088/api/v1/inventories/commit", "application/json", bytes.NewReader(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
