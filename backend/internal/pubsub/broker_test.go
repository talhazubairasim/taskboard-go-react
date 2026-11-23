package pubsub

import (
	"context"
	"testing"
	"time"
)

func TestPublishSubscribe(t *testing.T) {
	b := NewBroker()
	ctx, cancel := context.WithCancel(context.Background())
	ch1 := b.Subscribe(ctx)
	ch2 := b.Subscribe(ctx)
	if b.Subscribers() != 2 {
		t.Fatalf("subscribers = %d", b.Subscribers())
	}
	b.Publish(Event{Type: TaskDeleted, TaskID: "t1"})
	for _, ch := range []<-chan Event{ch1, ch2} {
		select {
		case e := <-ch:
			if e.TaskID != "t1" || e.Type != TaskDeleted {
				t.Fatalf("event = %+v", e)
			}
		case <-time.After(time.Second):
			t.Fatal("no event received")
		}
	}
	cancel()
	// channel is closed after cancellation
	select {
	case _, ok := <-ch1:
		if ok {
			t.Fatal("expected closed channel")
		}
	case <-time.After(time.Second):
		t.Fatal("channel not closed")
	}
	deadline := time.Now().Add(time.Second)
	for b.Subscribers() != 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if b.Subscribers() != 0 {
		t.Fatal("subscribers not removed")
	}
}

func TestSlowSubscriberDoesNotBlock(t *testing.T) {
	b := NewBroker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = b.Subscribe(ctx) // never read
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			b.Publish(Event{Type: TaskCreated})
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("publish blocked on a slow subscriber")
	}
}
