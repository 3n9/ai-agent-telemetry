package ulidgen

import (
	"sync"
	"testing"
)

const ulidCharset = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func isValidULIDChar(c byte) bool {
	for i := 0; i < len(ulidCharset); i++ {
		if ulidCharset[i] == c {
			return true
		}
	}
	return false
}

func TestNew_NonEmpty(t *testing.T) {
	id := New()
	if id == "" {
		t.Error("New() returned empty string")
	}
}

func TestNew_ValidULID(t *testing.T) {
	id := New()
	if len(id) != 26 {
		t.Errorf("expected ULID length 26, got %d: %q", len(id), id)
	}
	for i := 0; i < len(id); i++ {
		if !isValidULIDChar(id[i]) {
			t.Errorf("invalid ULID character %q at position %d in %q", id[i], i, id)
		}
	}
}

func TestNew_UniqueValues(t *testing.T) {
	a := New()
	b := New()
	if a == b {
		t.Errorf("two sequential New() calls returned identical value: %q", a)
	}
}

func TestNew_ConcurrentUnique(t *testing.T) {
	const n = 100
	results := make([]string, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			results[i] = New()
		}()
	}
	wg.Wait()

	seen := make(map[string]bool, n)
	for _, id := range results {
		if seen[id] {
			t.Errorf("duplicate ULID found in concurrent results: %q", id)
		}
		seen[id] = true
	}
}
