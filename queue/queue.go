package internal

import (
	"context"
	"sync"
	"time"
)

// Queue is a thread-safe implementation of a queue.
type Queue[T any] struct {
	el     []T
	mu     sync.RWMutex
	ch     chan T
	closed bool
}

// New creates a new empty queue.
func New[T any]() *Queue[T] {
	return NewFrom[T]()
}

// NewFrom creates a new queue prepopulated with these values.
//
// The values will be copied into the queue so modifications to the slice will not affect the queue. External
// modifications to the elements themselves still affect the in-queue elements.
func NewFrom[T any](args ...T) *Queue[T] {
	el := make([]T, len(args))
	copy(el, args)

	return &Queue[T]{
		el: el,
		mu: sync.RWMutex{},
		ch: make(chan T),
	}
}

// Close closes the queue and prevents new entries being added.
//
// Subsequent Add will panic for simplicity. Take can still be called to drain the queue.
func (q *Queue[T]) Close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
}

// IsClosed returns true if the queue has been closed and can no longer be added to.
//
// Take can still be called to drain the queue.
func (q *Queue[T]) IsClosed() bool {
	q.mu.RLock()
	v := q.closed
	q.mu.RUnlock()
	return v
}

// Add adds the file to the end of the queue.
//
// Add panics if the queue has been closed with Close. Add never blocks.
func (q *Queue[T]) Add(v T) {
	q.mu.RLock()
	closed := q.closed
	q.mu.RUnlock()
	if closed {
		panic("queue is closed")
	}

	// using the channel doesn't need mutex because "technically" the queue is never modified.
	// the sender and the receiver exchange the value directly without going through the queue. this pattern is also
	// used in other methods to facilitate direct exchange that skips blocking.
	select {
	case q.ch <- v:
	default:
		q.mu.Lock()
		q.el = append(q.el, v)
		q.mu.Unlock()
	}
}

// TryAdd attempts to add the file to the end of the queue.
//
// TryAdd will return false if the queue has been closed with Close. TryAdd never blocks.
func (q *Queue[T]) TryAdd(v T) bool {
	q.mu.RLock()
	closed := q.closed
	q.mu.RUnlock()
	if closed {
		return false
	}

	select {
	case q.ch <- v:
	default:
		q.mu.Lock()
		q.el = append(q.el, v)
		q.mu.Unlock()
	}
	return true
}

// Take blocks until an element can be retrieved from the front of the queue.
//
// The boolean return value is false if queue is empty after the context is done.
//
// Usage:
//
//	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
//	defer cancel()
//	v, ok := qu.Take(ctx)
//
// See TakeWithTimeout for a convenient method using the pattern above.
func (q *Queue[T]) Take(ctx context.Context) (v T, ok bool) {
	q.mu.Lock()
	closed := q.closed
	n := len(q.el)
	if n > 0 {
		v, q.el = q.el[0], q.el[1:]
		q.mu.Unlock()
		return v, true
	}
	q.mu.Unlock()

	if closed {
		return v, false
	}

	select {
	case <-ctx.Done():
		return v, false
	case v = <-q.ch:
		return v, true
	}
}

// TakeWithTimeout is a specialisation of Take that uses a derived context with the specified timeout duration.
//
// The boolean return value is false if queue is empty after timeout has expired.
func (q *Queue[T]) TakeWithTimeout(parent context.Context, timeout time.Duration) (v T, ok bool) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	return q.Take(ctx)
}

// TryTake retrieves an element from the front of the queue without blocking.
//
// The boolean return value is false if queue is empty at the time invocation is made.
func (q *Queue[T]) TryTake() (v T, ok bool) {
	select {
	case v, ok = <-q.ch:
	default:
	}
	return
}

// Size returns the current size of the queue.
func (q *Queue[T]) Size() int {
	q.mu.RLock()
	n := len(q.el)
	q.mu.RUnlock()
	return n
}
