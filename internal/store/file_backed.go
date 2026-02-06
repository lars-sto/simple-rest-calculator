package store

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"calculator-service/internal/model"
)

type FileBackedStore struct {
	mem *MemoryStore

	mu   sync.Mutex // protects file appends
	path string
	f    *os.File
	enc  *json.Encoder
}

// NewFileBackedStore loads existing entries from path (if present) and appends new entries to it
func NewFileBackedStore(path string, capacity int) (*FileBackedStore, error) {
	if path == "" {
		return nil, errors.New("path must not be empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	mem := NewMemoryStore(capacity)

	if err := loadJSONLIntoMemory(path, mem); err != nil {
		return nil, err
	}

	if err := compactJSONL(path, mem, capacity); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return &FileBackedStore{
		mem:  mem,
		path: path,
		f:    f,
		enc:  json.NewEncoder(f),
	}, nil
}

func (s *FileBackedStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.f != nil {
		return s.f.Close()
	}
	return nil
}

func (s *FileBackedStore) Add(ctx context.Context, e model.ResultEntry) error {
	if err := s.mem.Add(ctx, e); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.enc.Encode(e); err != nil {
		return err
	}

	return nil
}

func (s *FileBackedStore) ListRecent(ctx context.Context, n int) ([]model.ResultEntry, error) {
	return s.mem.ListRecent(ctx, n)
}

func loadJSONLIntoMemory(path string, mem *MemoryStore) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no history yet
		}
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e model.ResultEntry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		_ = mem.Add(context.Background(), e)
	}

	return sc.Err()
}

func compactJSONL(path string, mem *MemoryStore, capacity int) error {
	// If file doesn't exist, nothing to compact
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Read the last `capacity` entries from memory (already loaded from file)
	entries, err := mem.ListRecent(context.Background(), capacity)
	if err != nil {
		return err
	}

	// ListRecent returns newest-first, we want oldest-first in the compacted file
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	tmp := path + ".tmp"

	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			_ = f.Close()
			return err
		}
	}

	if err := f.Close(); err != nil {
		return err
	}

	_ = os.Remove(path)
	return os.Rename(tmp, path)
}
