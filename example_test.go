package messagebus_test

import (
	"context"
	"fmt"
	"time"

	messagebus "github.com/vardius/message-bus/v2"
)

type OrderPlaced struct {
	ID    int
	Total int // cents
}

func Example() {
	ctx := context.Background()
	bus := messagebus.New[OrderPlaced](100)

	_, err := bus.Subscribe("orders", func(e OrderPlaced) {
		fmt.Printf("receipt for order %d: $%d.%02d\n", e.ID, e.Total/100, e.Total%100)
	})
	if err != nil {
		panic(err)
	}

	if err := bus.Publish(ctx, "orders", OrderPlaced{ID: 7, Total: 1250}); err != nil {
		panic(err)
	}

	// Close waits for the handler to finish.
	if err := bus.Close(ctx); err != nil {
		panic(err)
	}
	// Output: receipt for order 7: $12.50
}

func ExampleBus_Subscribe() {
	ctx := context.Background()
	bus := messagebus.New[int](10)

	unsubscribe, err := bus.Subscribe("ticks", func(n int) { fmt.Println("tick", n) })
	if err != nil {
		panic(err)
	}

	_ = bus.Publish(ctx, "ticks", 1)
	_ = bus.Publish(ctx, "ticks", 2)
	unsubscribe()                    // 1 and 2 are already queued, so they are still handled
	_ = bus.Publish(ctx, "ticks", 3) // no subscribers left: dropped

	_ = bus.Close(ctx)
	// Output:
	// tick 1
	// tick 2
}

func ExampleBus_Publish() {
	ctx := context.Background()
	bus := messagebus.New[string](1)

	release := make(chan struct{})
	_, _ = bus.Subscribe("jobs", func(job string) {
		<-release
		fmt.Println("done", job)
	})

	_ = bus.Publish(ctx, "jobs", "a") // the handler takes a, and waits
	_ = bus.Publish(ctx, "jobs", "b") // b fills the queue

	// No room for c: give up after 10ms.
	timeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	err := bus.Publish(timeout, "jobs", "c")
	fmt.Println(err)

	close(release)
	_ = bus.Close(ctx)
	// Output:
	// messagebus: publish to "jobs": context deadline exceeded
	// done a
	// done b
}

func ExampleBus_Close() {
	ctx := context.Background()
	bus := messagebus.New[string](10)

	_, _ = bus.Subscribe("emails", func(to string) { fmt.Println("sent to", to) })
	for _, to := range []string{"ana", "ben", "cy"} {
		_ = bus.Publish(ctx, "emails", to)
	}

	// Stop the publishers first, then Close: it returns once every queued message has been handled.
	shutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := bus.Close(shutdown); err != nil {
		fmt.Println("some emails weren't sent:", err)
	}

	fmt.Println(bus.Publish(ctx, "emails", "dee"))
	// Output:
	// sent to ana
	// sent to ben
	// sent to cy
	// messagebus: bus closed
}
