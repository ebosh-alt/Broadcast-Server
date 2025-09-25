package events

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Subscription[T any] struct {
	ID int64
	C  <-chan T
}

type Subject[T any] struct {
	mu       sync.RWMutex
	nextID   int64
	subs     map[int64]chan T
	buf      int
	sendWait time.Duration
}

type Option func(*Subject[any])

func NewSubject[T any](buf int, sendWait time.Duration) *Subject[T] {
	return &Subject[T]{
		subs:     make(map[int64]chan T),
		buf:      buf,
		sendWait: sendWait,
	}
}

func (s *Subject[T]) Subscribe() Subscription[T] {
	ch := make(chan T, s.buf)
	id := atomic.AddInt64(&s.nextID, 1)
	s.mu.Lock()
	s.subs[id] = ch
	s.mu.Unlock()
	return Subscription[T]{ID: id, C: ch}
}

func (s *Subject[T]) Unsubscribe(id int64) {
	s.mu.Lock()
	if ch, ok := s.subs[id]; ok {
		delete(s.subs, id)
		close(ch)
	}
	s.mu.Unlock()
}

func (s *Subject[T]) Publish(ctx context.Context, v T) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, ch := range s.subs {
		if s.sendWait == 0 {
			select {
			case ch <- v:
			default:
				go s.Unsubscribe(id)
			}
			continue
		}
		timer := time.NewTimer(s.sendWait)
		select {
		case ch <- v:
			timer.Stop()
		case <-timer.C:
			go s.Unsubscribe(id)
		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
}

func (s *Subject[T]) Close() {
	s.mu.Lock()
	for id, ch := range s.subs {
		close(ch)
		delete(s.subs, id)
	}
	s.mu.Unlock()
}
