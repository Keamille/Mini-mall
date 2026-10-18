package order

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var notifyClient = &http.Client{Timeout: 2 * time.Second}

func notifyOrder(eventType, orderID, userID string, payload map[string]any) {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["order_id"] = orderID
	payload["user_id"] = userID
	body := map[string]any{
		"event_id":   orderID + ":" + eventType,
		"event_type": eventType,
		"user_id":    userID,
		"payload":    payload,
	}
	raw, _ := json.Marshal(body)
	resp, err := notifyClient.Post("http://127.0.0.1:8087/api/v1/notifications/events", "application/json", bytes.NewReader(raw))
	if err != nil {
		log.Printf("notify %s %s: %v", eventType, orderID, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("notify %s %s: status %d", eventType, orderID, resp.StatusCode)
	}
}
