package secrets

import (
	"context"
	"sync"
)

// serialQueue admits requests in enqueue order. A cancelled waiter leaves its
// place only after its predecessor completes, so later requests cannot overlap.
type serialQueue struct {
	mu   sync.Mutex
	tail <-chan struct{}
}

func (q *serialQueue) enter(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	q.mu.Lock()
	previous := q.tail
	done := make(chan struct{})
	q.tail = done
	q.mu.Unlock()
	if previous != nil {
		select {
		case <-previous:
		case <-ctx.Done():
			go func() { <-previous; close(done) }()
			return nil, ctx.Err()
		}
	}
	if err := ctx.Err(); err != nil {
		close(done)
		return nil, err
	}
	return func() { close(done) }, nil
}
