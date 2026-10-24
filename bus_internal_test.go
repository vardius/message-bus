package messagebus

import (
	"fmt"
	"testing"
)

// topicCount counts the topics the bus keeps an entry for.
func (b *Bus[T]) topicCount() int {
	n := 0
	for i := range b.shards {
		sh := &b.shards[i]
		sh.mu.RLock()
		n += len(sh.topics)
		sh.mu.RUnlock()
	}
	return n
}

// A bus that sees many short-lived topics must not keep an entry for each.
func TestUnsubscribeForgetsEmptyTopics(t *testing.T) {
	bus := New[int](0)
	var unsubscribes []func()
	for i := range 200 {
		for range 2 {
			unsubscribe, err := bus.Subscribe(fmt.Sprint("topic ", i), func(int) {})
			if err != nil {
				t.Fatal(err)
			}
			unsubscribes = append(unsubscribes, unsubscribe)
		}
	}

	for i, unsubscribe := range unsubscribes {
		unsubscribe()
		// Every second call ends the topic's last subscription.
		if got, want := bus.topicCount(), 200-(i+1)/2; got != want {
			t.Fatalf("after %d unsubscribes the bus keeps %d topics, want %d", i+1, got, want)
		}
	}
	if err := bus.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
}
