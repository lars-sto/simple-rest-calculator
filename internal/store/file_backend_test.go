package store

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"calculator-service/internal/model"
)

func TestFileBackedStore_LoadAndAppendAndCompact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recent.jsonl")

	s, err := NewFileBackedStore(path, 20)
	if err != nil {
		t.Fatalf("NewFileBackedStore error: %v", err)
	}

	// Add 25 entries
	for i := float64(1); i <= 25; i++ {
		v := i
		if err := s.Add(context.Background(), model.ResultEntry{
			Time:      time.Now(),
			Op:        model.OpAdd,
			A:         i,
			B:         0,
			Result:    &v,
			ExprHuman: "x",
		}); err != nil {
			t.Fatalf("Add error: %v", err)
		}
	}
	_ = s.Close()

	// Reopen: should load + compact to last 20 entries
	s2, err := NewFileBackedStore(path, 20)
	if err != nil {
		t.Fatalf("reopen NewFileBackedStore error: %v", err)
	}
	defer func() { _ = s2.Close() }()

	got, err := s2.ListRecent(context.Background(), 20)
	if err != nil {
		t.Fatalf("ListRecent error: %v", err)
	}
	if len(got) != 20 {
		t.Fatalf("expected 20 entries, got %d", len(got))
	}

	// newest first: 25..6
	if got[0].A != 25 {
		t.Fatalf("expected newest A=25, got %.5f", got[0].A)
	}
	if got[len(got)-1].A != 6 {
		t.Fatalf("expected oldest kept A=6, got %.5f", got[len(got)-1].A)
	}

	// File should be compacted to 20 lines
	lines, err := countNonEmptyLines(path)
	if err != nil {
		t.Fatalf("countNonEmptyLines error: %v", err)
	}
	if lines != 20 {
		t.Fatalf("expected 20 lines after compaction, got %d", lines)
	}
}

func TestFileBackedStore_SkipsMalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recent.jsonl")

	// Write a malformed line + a valid JSON line
	err := os.WriteFile(path, []byte("NOT JSON\n{\"time\":\"2020-01-01T00:00:00Z\",\"op\":\"add\",\"a\":1,\"b\":2,\"result\":3,\"exprHuman\":\"1 + 2 = 3\"}\n"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	s, err := NewFileBackedStore(path, 20)
	if err != nil {
		t.Fatalf("NewFileBackedStore error: %v", err)
	}
	defer func() { _ = s.Close() }()

	got, err := s.ListRecent(context.Background(), 20)
	if err != nil {
		t.Fatalf("ListRecent error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 valid entry loaded, got %d", len(got))
	}
	if got[0].A != 1 || got[0].B != 2 {
		t.Fatalf("unexpected entry: A=%.5f B=%.5f", got[0].A, got[0].B)
	}

	// After compaction-on-start, file should now only contain the valid entry (1 line)
	lines, err := countNonEmptyLines(path)
	if err != nil {
		t.Fatalf("countNonEmptyLines error: %v", err)
	}
	if lines != 1 {
		t.Fatalf("expected 1 line after compaction, got %d", lines)
	}
}

func countNonEmptyLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = f.Close()
	}()

	sc := bufio.NewScanner(f)
	count := 0
	for sc.Scan() {
		if len(sc.Bytes()) > 0 {
			count++
		}
	}
	return count, sc.Err()
}
