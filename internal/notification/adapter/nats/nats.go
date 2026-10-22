package natsadp

import (
	"context"
	"encoding/json"
	"github.com/your-org/go-minimall/internal/notification/app"
	"log"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

// Same JSON shape as order.EventEnvelope; do not import internal/order.
type eventEnvelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Producer   string          `json:"producer"`
	Data       json.RawMessage `json:"data"`
}

// StartOrderCreatedConsumer pulls minimall.order.order_created from JetStream.
func StartOrderCreatedConsumer(h *app.Handler, js jetstream.JetStream) error {
	ctx := context.Background()

	if _, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "ORDERS",
		Subjects: []string{"minimall.order.>"},
	}); err != nil {
		return err
	}

	cons, err := js.CreateOrUpdateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
		Durable:       "notify-worker",
		FilterSubject: "minimall.order.order_created",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	_, err = cons.Consume(func(msg jetstream.Msg) {
		var env eventEnvelope
		if err := json.Unmarshal(msg.Data(), &env); err != nil {
			log.Printf("nats bad envelope: %v", err)
			_ = msg.Term()
			return
		}

		var data map[string]any
		_ = json.Unmarshal(env.Data, &data)
		userID, _ := data["user_id"].(string)
		if userID == "" || env.EventID == "" || env.EventType == "" {
			log.Printf("nats missing fields: id=%s type=%s user=%s", env.EventID, env.EventType, userID)
			_ = msg.Term()
			return
		}

		if _, err := h.Repo.Record(env.EventID, env.EventType, userID, env.Data); err != nil {
			log.Printf("nats record %s: %v", env.EventID, err)
			return
		}
		if err := msg.Ack(); err != nil {
			log.Printf("nats ack %s: %v", env.EventID, err)
		}
	})
	return err
}
