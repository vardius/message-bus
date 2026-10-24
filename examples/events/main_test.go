package main

import (
	"maps"
	"strings"
	"testing"
)

func TestPlaceOrders(t *testing.T) {
	var receipts strings.Builder
	stock := map[string]int{"tea": 10, "cake": 4}

	err := placeOrders(t.Context(), &receipts, stock, []OrderPlaced{
		{ID: 1, Item: "tea", Quantity: 2},
		{ID: 2, Item: "cake", Quantity: 1},
		{ID: 3, Item: "tea", Quantity: 3},
	})
	if err != nil {
		t.Fatalf("placeOrders() error = %v", err)
	}

	if want := "order 1: 2 × tea\norder 2: 1 × cake\norder 3: 3 × tea\n"; receipts.String() != want {
		t.Errorf("receipts = %q, want %q", receipts.String(), want)
	}
	if want := map[string]int{"tea": 5, "cake": 3}; !maps.Equal(stock, want) {
		t.Errorf("stock = %v, want %v", stock, want)
	}
}
