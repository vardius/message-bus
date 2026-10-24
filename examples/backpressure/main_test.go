package main

import (
	"context"
	"errors"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	messagebus "github.com/vardius/message-bus/v2"
)

func TestPublishAll(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](2)
		var done []int
		if _, err := bus.Subscribe("jobs", func(job int) {
			time.Sleep(100 * time.Millisecond)
			done = append(done, job)
		}); err != nil {
			t.Fatal(err)
		}

		// 1 is taken at once and 2 and 3 fill the queue. 4, 5 and 6 each wait 30ms in vain;
		// 7 gets in when 1 is done, at 100ms, and 8 waits in vain again.
		skipped, err := publishAll(t.Context(), bus, []int{1, 2, 3, 4, 5, 6, 7, 8}, 30*time.Millisecond)
		if err != nil {
			t.Fatalf("publishAll() error = %v", err)
		}
		if want := []int{4, 5, 6, 8}; !slices.Equal(skipped, want) {
			t.Errorf("skipped %v, want %v", skipped, want)
		}

		if err := bus.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
		if want := []int{1, 2, 3, 7}; !slices.Equal(done, want) {
			t.Errorf("done %v, want %v", done, want)
		}
	})
}

func TestPublishAllCancelled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		bus := messagebus.New[int](0)
		if _, err := bus.Subscribe("jobs", func(int) { time.Sleep(time.Second) }); err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		_, err := publishAll(ctx, bus, []int{1, 2}, time.Minute)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("publishAll() error = %v, want %v", err, context.Canceled)
		}

		if err := bus.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
}
