package order

import (
	"context"
	"encoding/json"
	"log"
	"time"

	notificationv1 "github.com/your-org/go-minimall/api/gen/notification/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


func notifyOrder(eventType, orderID, userID string, payload map[string]any) {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["order_id"] = orderID
	payload["user_id"] = userID
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("notify %s %s: %v", eventType, orderID, err)
		return
	}
	conn, err := grpc.NewClient("127.0.0.1:9009", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("notify %s %s: %v", eventType, orderID, err)
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = notificationv1.NewNotificationServiceClient(conn).RecordEvent(ctx, &notificationv1.RecordEventRequest{
		EventId:     orderID + ":" + eventType,
		EventType:   eventType,
		UserId:      userID,
		PayloadJson: string(raw),
	})
	if err != nil {
		log.Printf("notify %s %s: %v", eventType, orderID, err)
	}
}
