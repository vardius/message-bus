package message_bus

import (
	"fmt"
	"reflect"
	"sync"
)

type MessageBus interface {
	Publish(topic string, args ...interface{})
	Subscribe(topic string, fn interface{}) error
	Unsubscribe(topic string, fn interface{}) error
}

type handlersMap map[string][]reflect.Value

type messageBus struct {
	mtx      sync.RWMutex
	handlers handlersMap
}

func (b *messageBus) Publish(topic string, args ...interface{}) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()

	if hs, ok := b.handlers[topic]; ok {
		fArgs := make([]reflect.Value, 0)
		for _, arg := range args {
			fArgs = append(fArgs, reflect.ValueOf(arg))
		}

		for _, h := range hs {
			go h.Call(fArgs)
		}
	}
}

func (b *messageBus) Subscribe(topic string, fn interface{}) error {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(fn))
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()

	b.handlers[topic] = append(b.handlers[topic], reflect.ValueOf(fn))

	return nil
}

func (b *messageBus) Unsubscribe(topic string, fn interface{}) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	if _, ok := b.handlers[topic]; ok {
		rv := reflect.ValueOf(fn)

		for i, h := range b.handlers[topic] {
			if h == rv {
				b.handlers[topic] = append(b.handlers[topic][:i], b.handlers[topic][i+1:]...)
			}
		}

		return nil
	}

	return fmt.Errorf("Topic %s doesn't exist", topic)
}

func New() MessageBus {
	return &messageBus{
		handlers: make(handlersMap),
	}
}
