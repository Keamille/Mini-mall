package order

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// EventEnvelope 是订单发到 NATS 的统一外壳。
// notification 那边也要用同样的 JSON 字段名去解。
type EventEnvelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Producer   string          `json:"producer"`
	Data       json.RawMessage `json:"data"`
}

// publishOrderEvent 把「订单上发生了一件事」发到 NATS。
// 失败只打日志，不返回 error，避免拖垮下单。
func (h *Handler) publishOrderEvent(eventType, orderID, userID string, payload map[string]any) {
	if h.js == nil {
		return
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["order_id"] = orderID
	payload["user_id"] = userID

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("nats marshal data %s %s: %v", eventType, orderID, err)
		return
	}

	env := EventEnvelope{
		EventID:    orderID + ":" + eventType, // 和 gRPC 的 event_id 保持一致，方便幂等
		EventType:  eventType,                 // 如 order_created，给 record/templateOf 用
		OccurredAt: time.Now().UTC(),
		Producer:   "order-svc",
		Data:       data,
	}
	raw, err := json.Marshal(env)
	if err != nil {
		log.Printf("nats marshal envelope %s %s: %v", eventType, orderID, err)
		return
	}

	// 注意：subject 必须是具体名字，不能写 minimall.order.>
	subject := "minimall.order." + eventType

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := h.js.Publish(ctx, subject, raw); err != nil {
		log.Printf("nats publish %s: %v", subject, err)
		return
	}
	log.Printf("nats published %s event_id=%s", subject, env.EventID)
}
