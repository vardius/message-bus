// Command backpressure publishes jobs faster than its handler can take them. When the queue is full,
// it waits a little for room, then skips the job rather than stall.
//
//	go run ./examples/backpressure
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	messagebus "github.com/vardius/message-bus/v2"
)

// publishAll publishes every job, waiting at most wait for room each time, and returns the jobs that didn't fit.
func publishAll(ctx context.Context, bus *messagebus.Bus[int], jobs []int, wait time.Duration) ([]int, error) {
	var skipped []int
	for _, job := range jobs {
		err := publishWithin(ctx, bus, job, wait)
		if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil {
			skipped = append(skipped, job)
			continue
		}
		if err != nil {
			return skipped, err
		}
	}
	return skipped, nil
}

func publishWithin(ctx context.Context, bus *messagebus.Bus[int], job int, wait time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	return bus.Publish(ctx, "jobs", job)
}

func main() {
	ctx := context.Background()
	// The handler can hold one job and queue two more.
	bus := messagebus.New[int](2)
	if _, err := bus.Subscribe("jobs", func(job int) {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("done", job)
	}); err != nil {
		log.Fatal(err)
	}

	skipped, err := publishAll(ctx, bus, []int{1, 2, 3, 4, 5, 6, 7, 8}, 30*time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("skipped", skipped)

	if err := bus.Close(ctx); err != nil {
		log.Fatal(err)
	}
}
