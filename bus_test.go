package messagebus

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	bus := New(runtime.NumCPU())

	if bus == nil {
		t.Fail()
	}
}

func TestSubscribe(t *testing.T) {
	bus := New(runtime.NumCPU())

	if bus.Subscribe("test", func() {}) != nil {
		t.Fail()
	}

	if bus.Subscribe("test", 2) == nil {
		t.Fail()
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := New(runtime.NumCPU())

	handler := func() {}

	bus.Subscribe("test", handler)

	if err := bus.Unsubscribe("test", handler); err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if err := bus.Unsubscribe("unexisted", func() {}); err == nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestClose(t *testing.T) {
	bus := New(runtime.NumCPU())

	handler := func() {}

	bus.Subscribe("test", handler)

	if 0 == len(original.handlers) {
		fmt.Println("Did not subscribed handler")
		t.Fail()
	}

	bus.Close()

	original, ok := bus.(*messageBus)
	if !ok {
		fmt.Println("Could not cast message bus to its original type")
		t.Fail()
	}

	if 0 != len(original.handlers) {
		fmt.Println("Did not unsubscribed all handlers")
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	bus := New(runtime.NumCPU())

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
	bus := New(runtime.NumCPU())
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
