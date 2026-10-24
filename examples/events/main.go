// Command events publishes orders to two handlers that don't know about each other:
// one writes receipts, the other keeps the stock.
//
//	go run ./examples/events
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	messagebus "github.com/vardius/message-bus/v2"
)

type OrderPlaced struct {
	ID       int
	Item     string
	Quantity int
}

// placeOrders publishes every order, writes a receipt for each, and takes what was ordered out of stock.
func placeOrders(ctx context.Context, receipts io.Writer, stock map[string]int, orders []OrderPlaced) error {
	bus := messagebus.New[OrderPlaced](100)

	// Each handler runs on its own goroutine, one order at a time, so stock needs no lock:
	// only this handler touches it until Close returns.
	if _, err := bus.Subscribe("orders", func(o OrderPlaced) {
		stock[o.Item] -= o.Quantity
	}); err != nil {
		return err
	}
	if _, err := bus.Subscribe("orders", func(o OrderPlaced) {
		fmt.Fprintf(receipts, "order %d: %d × %s\n", o.ID, o.Quantity, o.Item)
	}); err != nil {
		return err
	}

	for _, o := range orders {
		if err := bus.Publish(ctx, "orders", o); err != nil {
			return err
		}
	}

	// Close returns once both handlers have handled every order.
	return bus.Close(ctx)
}

func main() {
	stock := map[string]int{"tea": 10, "cake": 4}
	orders := []OrderPlaced{
		{ID: 1, Item: "tea", Quantity: 2},
		{ID: 2, Item: "cake", Quantity: 1},
		{ID: 3, Item: "tea", Quantity: 3},
	}

	if err := placeOrders(context.Background(), os.Stdout, stock, orders); err != nil {
		log.Fatal(err)
	}
	fmt.Println("left:", stock)
}
