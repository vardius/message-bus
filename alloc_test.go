package messagebus_test

import (
	"context"
	"testing"

	messagebus "github.com/vardius/message-bus/v2"
)

func noop(int) {}

// subscribed returns a bus with n subscribers to "t", each with a queue of queueSize messages.
// It is closed when the test ends.
func subscribed(tb testing.TB, n, queueSize int) *messagebus.Bus[int] {
	bus := messagebus.New[int](queueSize)
	for range n {
		if _, err := bus.Subscribe("t", noop); err != nil {
			tb.Fatal(err)
		}
	}
	tb.Cleanup(func() {
		if err := bus.Close(context.Background()); err != nil {
			tb.Error(err)
		}
	})
	return bus
}

// TestAllocations guards the allocations per call. Lower a number when an optimisation lands;
// never raise one without a reason in the commit.
//
// Subscribe isn't guarded: whether the runtime reuses an exited goroutine for the new one varies from run to run.
func TestAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector changes allocation counts")
	}

	tests := []struct {
		name      string
		subs      int
		queueSize int
		want      float64
	}{
		// The message is copied into each queue, and waking a waiting handler allocates nothing.
		// The queues fit more messages than AllocsPerRun publishes, so these never wait for room.
		{"Publish, no subscribers", 0, 1024, 0},
		{"Publish, 1 subscriber", 1, 1024, 0},
		{"Publish, 10 subscribers", 10, 1024, 0},
		// With no room, Publish waits for the handler on most calls. Background's Done is nil, so waiting
		// allocates nothing; a cancellable ctx would make its Done channel the first time.
		{"Publish waits for room, Background ctx", 1, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := subscribed(t, tt.subs, tt.queueSize)
			ctx := context.Background()
			msg := 0
			got := testing.AllocsPerRun(1000, func() {
				msg++
				_ = bus.Publish(ctx, "t", msg)
			})
			if got != tt.want {
				t.Errorf("allocates %v times per call, want %v", got, tt.want)
			}
		})
	}
}
