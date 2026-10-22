package app

import "github.com/your-org/go-minimall/internal/order/port"

type Handler struct {
	Repo      port.Repository
	Inventory port.Inventory
	Coupons   port.Coupons
	Notify    port.Notifier
	Events    port.Publisher
}

func NewHandler(repo port.Repository, inventory port.Inventory, coupons port.Coupons, notify port.Notifier, events port.Publisher) *Handler {
	return &Handler{Repo: repo, Inventory: inventory, Coupons: coupons, Notify: notify, Events: events}
}
