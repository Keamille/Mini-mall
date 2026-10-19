// One-off acceptance client for inventory gRPC ReserveStock (:9008).
//go:build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	inventoryv1 "github.com/your-org/go-minimall/api/gen/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const invHTTP = "http://127.0.0.1:8088"
const invGRPC = "127.0.0.1:9008"

func main() {
	fail := 0
	check := func(name string, ok bool, detail string) {
		if ok {
			fmt.Printf("[OK] %s\n", name)
			return
		}
		fmt.Printf("[FAIL] %s -- %s\n", name, detail)
		fail++
	}

	skuID := os.Getenv("SKU_ID")
	if skuID == "" {
		skuID = "01ACCEPT000000000000000010"
	}
	orderOK := "01ORDER0000000000000000010"
	orderBad := "01ORDERB000000000000000010"

	// Ensure stock exists (init is idempotent only on insert — use unique sku or POST init).
	initBody, _ := json.Marshal(map[string]any{"sku_id": skuID, "available": 50})
	resp, err := http.Post(invHTTP+"/api/v1/inventories", "application/json", bytes.NewReader(initBody))
	if err != nil {
		check("HTTP init stock", false, err.Error())
		os.Exit(1)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	initOK := resp.StatusCode == 200 || resp.StatusCode == 201
	if !initOK {
		got, err := http.Get(invHTTP + "/api/v1/inventories/" + skuID)
		if err == nil {
			var stock struct {
				Code int `json:"code"`
			}
			_ = json.NewDecoder(got.Body).Decode(&stock)
			got.Body.Close()
			initOK = stock.Code == 0
		}
	}
	check("HTTP init stock", initOK, string(body))

	conn, err := grpc.NewClient(invGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		check("grpc dial", false, err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	client := inventoryv1.NewInventoryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.ReserveStock(ctx, &inventoryv1.ReserveStockRequest{})
	check("InvalidArgument empty req", status.Code(err) == codes.InvalidArgument, fmt.Sprintf("%v", err))

	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	out, err := client.ReserveStock(ctx2, &inventoryv1.ReserveStockRequest{
		OrderId: orderOK,
		Items:   []*inventoryv1.ReserveItem{{SkuId: skuID, Quantity: 2}},
	})
	check("ReserveStock OK", err == nil && out.GetReservationId() != "", fmt.Sprintf("err=%v id=%q", err, out.GetReservationId()))
	firstID := out.GetReservationId()

	ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel3()
	out2, err := client.ReserveStock(ctx3, &inventoryv1.ReserveStockRequest{
		OrderId: orderOK,
		Items:   []*inventoryv1.ReserveItem{{SkuId: skuID, Quantity: 2}},
	})
	check("idempotent same order_id", err == nil && out2.GetReservationId() == firstID,
		fmt.Sprintf("first=%q second=%q err=%v", firstID, out2.GetReservationId(), err))

	got, err := http.Get(invHTTP + "/api/v1/inventories/" + skuID)
	if err != nil {
		check("GET stock", false, err.Error())
	} else {
		var stock struct {
			Code int `json:"code"`
			Data struct {
				Available int `json:"available"`
				Reserved  int `json:"reserved"`
			} `json:"data"`
		}
		_ = json.NewDecoder(got.Body).Decode(&stock)
		got.Body.Close()
		check("available 48", stock.Data.Available == 48, fmt.Sprintf("%d", stock.Data.Available))
		check("reserved 2", stock.Data.Reserved == 2, fmt.Sprintf("%d", stock.Data.Reserved))
	}

	ctx4, cancel4 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel4()
	_, err = client.ReserveStock(ctx4, &inventoryv1.ReserveStockRequest{
		OrderId: orderBad,
		Items:   []*inventoryv1.ReserveItem{{SkuId: skuID, Quantity: 9999}},
	})
	check("FailedPrecondition insufficient", status.Code(err) == codes.FailedPrecondition, fmt.Sprintf("%v", err))

	if fail > 0 {
		os.Exit(1)
	}
	fmt.Println("All gRPC reserve acceptance checks passed.")
}
