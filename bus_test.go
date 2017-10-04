package message_bus

import (
	"errors"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	bus := New()

	if bus == nil {
		t.Fail()
	}
}

func TestSubscribe(t *testing.T) {
	bus := New()

	if bus.Subscribe("test", func() {}) != nil {
		t.Fail()
	}

	if bus.Subscribe("test", 2) == nil {
		t.Fail()
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := New()

	handler := func() {}

	bus.Subscribe("test", handler)
	bus.Unsubscribe("test", handler)

	if bus.Unsubscribe("unexisted", func() {}) != nil {
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	bus := New()

	var wg sync.WaitGroup
	wg.Add(2)

	first := false
	second := false

	bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		first = v
	})

	bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		second = v
	})

	bus.Publish("topic", true)

	wg.Wait()

	if first == false || second == false {
		t.Fail()
	}
}

func TestHandleError(t *testing.T) {
	bus := New()
	bus.Subscribe("topic", func(out chan<- error) {
		out <- errors.New("I do throw error")
	})

	out := make(chan error)
	defer close(out)

	bus.Publish("topic", out)

	if <-out == nil {
		t.Fail()
	}
}
