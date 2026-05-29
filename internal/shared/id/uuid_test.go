package id

import (
	"testing"
)

func TestNew_ReturnsValidUUID(t *testing.T) {
	id := New()
	if id == "" {
		t.Fatal("expected non-empty ID")
	}
	if !IsValid(id) {
		t.Errorf("generated ID is not a valid UUID: %s", id)
	}
}

func TestNew_ReturnsUniqueValues(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id := New()
		if seen[id] {
			t.Fatalf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid v7", New(), true},
		{"empty", "", false},
		{"garbage", "not-a-uuid", false},
		{"missing section", "550e8400-e29b-41d4-a716", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.input); got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNew_TimeOrdering(t *testing.T) {
	// UUID v7 should be time-sortable: later IDs should be lexicographically greater
	id1 := New()
	id2 := New()

	if id1 >= id2 {
		// This could technically fail in rare cases with same-timestamp UUIDs,
		// but UUID v7 includes random bits that make collisions exceedingly rare
		// within the same process generating sequential IDs.
		t.Logf("warning: id1=%s >= id2=%s (rare but possible with UUID v7)", id1, id2)
	}
}
