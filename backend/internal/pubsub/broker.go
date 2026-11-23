// Package pubsub is a small in-process event broker used for real-time task updates.
package pubsub

import (
	"context"
	"sync"

	"taskboard/internal/models"
)

// EventType says what happened to a task.
type EventType string

// Event types.
const (
	TaskCreated EventType = "TASK_CREATED"
	TaskUpdated EventType = "TASK_UPDATED"
	TaskDeleted EventType = "TASK_DELETED"
)

// Event is published after every task change.
type Event struct {
	Type   EventType
	TaskID string
	Task   *models.TaskDetails // nil for TaskDeleted
}

// Broker fans events out to every active subscriber.
type Broker struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

// NewBroker builds an empty broker.
func NewBroker() *Broker {
	return &Broker{subs: make(map[chan Event]struct{})}
}

// Subscribe returns a channel that receives events until ctx is cancelled.
func (b *Broker) Subscribe(ctx context.Context) <-chan Event {
	ch := make(chan Event, 32)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subs, ch)
		close(ch)
		b.mu.Unlock()
	}()
	return ch
}

// Publish delivers e to every subscriber. Slow subscribers whose buffer is
// full miss the event rather than blocking the publisher.
func (b *Broker) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		select {
		case ch <- e:
		default:
		}
	}
}

// Subscribers returns the number of active subscribers.
func (b *Broker) Subscribers() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}
