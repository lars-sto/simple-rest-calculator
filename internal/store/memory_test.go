package store

import (
	"context"
	"testing"
	"time"

	"calculator-service/internal/model"
)

func TestMemoryStore_Empty(t *testing.T) {
	s := NewMemoryStore(20)

	got, err := s.ListRecent(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListRecent returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d entries", len(got))
	}
}

func TestMemoryStore_OrderNewestFirst(t *testing.T) {
	s := NewMemoryStore(20)

	for i := float64(1); i <= 3; i++ {
		v := i
		_ = s.Add(context.Background(), model.ResultEntry{
			Time:      time.Now(),
			Op:        model.OpAdd,
			A:         i,
			B:         0,
			Result:    &v,
			ExprHuman: "x",
		})
	}

	got, err := s.ListRecent(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListRecent returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}

	// newest first
	if got[0].A != 3 || got[1].A != 2 {
		t.Fatalf("expected [3,2], got [%.5f,%.5f]", got[0].A, got[1].A)
	}
}

func TestMemoryStore_OverwriteAfterCapacity(t *testing.T) {
	s := NewMemoryStore(20)

	// Add 25 entries, capacity is 20 -> we should keep entries 6..25 (newest first 25..6)
	for i := float64(1); i <= 25; i++ {
		v := i
		_ = s.Add(context.Background(), model.ResultEntry{
			Time:      time.Now(),
			Op:        model.OpAdd,
			A:         i,
			B:         0,
			Result:    &v,
			ExprHuman: "x",
		})
	}

	got, err := s.ListRecent(context.Background(), 20)
	if err != nil {
		t.Fatalf("ListRecent returned error: %v", err)
	}
	if len(got) != 20 {
		t.Fatalf("expected 20 entries, got %d", len(got))
	}

	if got[0].A != 25 {
		t.Fatalf("expected newest A=25, got %.5f", got[0].A)
	}
	if got[len(got)-1].A != 6 {
		t.Fatalf("expected oldest kept A=6, got %.5f", got[len(got)-1].A)
	}
}

func TestMemoryStore_ListRecentClampsAndDefaults(t *testing.T) {
	s := NewMemoryStore(20)

	for i := float64(1); i <= 3; i++ {
		v := i
		_ = s.Add(context.Background(), model.ResultEntry{
			Time:      time.Now(),
			Op:        model.OpAdd,
			A:         i,
			B:         0,
			Result:    &v,
			ExprHuman: "x",
		})
	}

	// n bigger than size -> clamp to size
	got, _ := s.ListRecent(context.Background(), 999)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}

	// n <= 0 -> default 5, clamped to size (3)
	got2, _ := s.ListRecent(context.Background(), 0)
	if len(got2) != 3 {
		t.Fatalf("expected 3 entries (default 5 clamped), got %d", len(got2))
	}
}
