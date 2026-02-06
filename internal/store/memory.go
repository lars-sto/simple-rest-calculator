package store

import (
	"context"
	"sync"

	"calculator-service/internal/model"
)

type MemoryStore struct {
	mu   sync.RWMutex
	cap  int
	buf  []model.ResultEntry
	next int
	size int
}

func NewMemoryStore(capacity int) *MemoryStore {
	if capacity <= 0 {
		capacity = 20
	}
	return &MemoryStore{
		cap: capacity,
		buf: make([]model.ResultEntry, capacity),
	}
}

func (s *MemoryStore) Add(ctx context.Context, e model.ResultEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buf[s.next] = e
	s.next = (s.next + 1) % s.cap
	if s.size < s.cap {
		s.size++
	}
	return nil
}

func (s *MemoryStore) ListRecent(ctx context.Context, n int) ([]model.ResultEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n <= 0 {
		n = 5
	}
	if n > s.size {
		n = s.size
	}

	out := make([]model.ResultEntry, 0, n)

	idx := s.next - 1
	if idx < 0 {
		idx = s.cap - 1
	}

	for i := 0; i < n; i++ {
		out = append(out, s.buf[idx])
		idx--
		if idx < 0 {
			idx = s.cap - 1
		}
	}
	return out, nil
}
