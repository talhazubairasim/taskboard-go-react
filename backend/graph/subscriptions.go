package graph

import (
	"context"

	"taskboard/graph/model"
	"taskboard/internal/pubsub"
)

// taskStream forwards matching task events until ctx ends.
func (r *subscriptionResolver) taskStream(ctx context.Context, typ pubsub.EventType, onlyID *string) <-chan *model.Task {
	events := r.Broker.Subscribe(ctx)
	out := make(chan *model.Task, 1)
	go func() {
		defer close(out)
		for e := range events {
			if e.Type != typ || e.Task == nil || (onlyID != nil && *onlyID != e.TaskID) {
				continue
			}
			select {
			case out <- toTask(e.Task):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
