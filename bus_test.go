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

	if err := bus.Subscribe("test", handler); err != nil {
		t.Fatal(err)
	}

	if err := bus.Unsubscribe("test", handler); err != nil {
		fmt.Println(err)
		t.Fail()
	}

	if err := bus.Unsubscribe("non-existed", func() {}); err == nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestClose(t *testing.T) {
	bus := New(runtime.NumCPU())

	handler := func() {}

	if err := bus.Subscribe("test", handler); err != nil {
		t.Fatal(err)
	}

	original, ok := bus.(*messageBus)
	if !ok {
		fmt.Println("Could not cast message bus to its original type")
		t.Fail()
	}

	if 0 == len(original.handlers) {
		fmt.Println("Did not subscribed handler to topic")
		t.Fail()
	}

	bus.Close("test")

	if 0 != len(original.handlers) {
		fmt.Println("Did not unsubscribed handlers from topic")
		t.Fail()
	}
}

func TestPublish(t *testing.T) {
	bus := New(runtime.NumCPU())

	var wg sync.WaitGroup
	wg.Add(2)

	first := false
	second := false

	if err := bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		first = v
	}); err != nil {
		t.Fatal(err)
	}

	if err := bus.Subscribe("topic", func(v bool) {
		defer wg.Done()
		second = v
	}); err != nil {
		t.Fatal(err)
	}

	bus.Publish("topic", true)

	wg.Wait()

	if first == false || second == false {
		t.Fail()
	}
}

func TestHandleError(t *testing.T) {
	bus := New(runtime.NumCPU())
	if err := bus.Subscribe("topic", func(out chan<- error) {
		out <- errors.New("I do throw error")
	}); err != nil {
		t.Fatal(err)
	}

	out := make(chan error)
	defer close(out)

	bus.Publish("topic", out)

	if <-out == nil {
		t.Fail()
	}
}
