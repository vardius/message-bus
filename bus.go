package messagebus

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// MessageBus implements publish/subscribe messaging paradigm
type MessageBus interface {
	Publish(topic string, args ...interface{})
	Subscribe(topic string, fn interface{}) error
	Unsubscribe(topic string, fn interface{}) error
}

type handlersMap map[string][]*handler

type handler struct {
	ctx      context.Context
	callback reflect.Value
	cancel   context.CancelFunc
	queue    chan []reflect.Value
}

type messageBus struct {
	maxConcurrentCalls int
	mtx                sync.RWMutex
	handlers           handlersMap
}

// Publish publishes arguments to the given topic subscribers
func (b *messageBus) Publish(topic string, args ...interface{}) {
	b.mtx.RLock()
	defer b.mtx.RUnlock()

	if hs, ok := b.handlers[topic]; ok {
		rArgs := buildHandlerArgs(args)

		for _, h := range hs {
			h.queue <- rArgs
		}
	}
}

// Subscribe subscribes to the given topic
func (b *messageBus) Subscribe(topic string, fn interface{}) error {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		return fmt.Errorf("%s is not a reflect.Func", reflect.TypeOf(fn))
	}

	ctx, cancel := context.WithCancel(context.Background())

	h := &handler{
		callback: reflect.ValueOf(fn),
		ctx:      ctx,
		cancel:   cancel,
		queue:    make(chan []reflect.Value, b.maxConcurrentCalls),
	}

	go func() {
		for {
			select {
			case args, ok := <-h.queue:
				if ok {
					h.callback.Call(args)
				}
			case <-h.ctx.Done():
				return
			}
		}
	}()

	b.mtx.Lock()
	defer b.mtx.Unlock()

	b.handlers[topic] = append(b.handlers[topic], h)

	return nil
}

// Unsubscribe unsubscribes from the given topic
func (b *messageBus) Unsubscribe(topic string, fn interface{}) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	if _, ok := b.handlers[topic]; ok {
		rv := reflect.ValueOf(fn)

		for i, h := range b.handlers[topic] {
			if h.callback == rv {
				h.cancel()
				close(h.queue)
				b.handlers[topic] = append(b.handlers[topic][:i], b.handlers[topic][i+1:]...)
			}
		}

		return nil
	}

	return fmt.Errorf("Topic %s doesn't exist", topic)
}

// Close unsubscribes all handlers
func (b *messageBus) Close() {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	for _, handlers := range b.handlers {
		for _, h := range handlers {
			h.cancel()
			close(h.queue)
		}
	}

	b.handlers = make(handlersMap)
}

func buildHandlerArgs(args []interface{}) []reflect.Value {
	reflectedArgs := make([]reflect.Value, 0)

	for _, arg := range args {
		reflectedArgs = append(reflectedArgs, reflect.ValueOf(arg))
	}

	return reflectedArgs
}

// New creates new MessageBus
// maxConcurrentCalls limits concurrency by using a buffered channel semaphore
func New(maxConcurrentCalls int) MessageBus {
	return &messageBus{
		maxConcurrentCalls: maxConcurrentCalls,
		handlers:           make(handlersMap),
	}
}
