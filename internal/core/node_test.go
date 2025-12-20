package core

import (
	"testing"
)

func TestNewLeaf(t *testing.T) {
	s := "hello"
	l := NewLeaf(s)
	if l == nil {
		t.Fatal("NewLeaf returned nil")
	}
	if l.val != s {
		t.Errorf("expected val %q, got %q", s, l.val)
	}
}

func TestLeaf_Len(t *testing.T) {
	s := "hello world"
	l := NewLeaf(s)
	if got := l.Len(); got != len(s) {
		t.Errorf("Len() = %d, want %d", got, len(s))
	}
}

func TestLeaf_ByteAt(t *testing.T) {
	s := "abc"
	l := NewLeaf(s)
	if b := l.ByteAt(0); b != 'a' {
		t.Errorf("ByteAt(0) = %c, want 'a'", b)
	}
	if b := l.ByteAt(1); b != 'b' {
		t.Errorf("ByteAt(1) = %c, want 'b'", b)
	}
	if b := l.ByteAt(2); b != 'c' {
		t.Errorf("ByteAt(2) = %c, want 'c'", b)
	}
}

func TestLeaf_String(t *testing.T) {
	s := "test string"
	l := NewLeaf(s)
	if got := l.String(); got != s {
		t.Errorf("String() = %q, want %q", got, s)
	}
}

func TestLeaf_Depth(t *testing.T) {
	l := NewLeaf("leaf")
	if d := l.Depth(); d != 0 {
		t.Errorf("Depth() = %d, want 0", d)
	}
}

func TestLeaf_Slice(t *testing.T) {
	s := "012345"
	l := NewLeaf(s)

	tests := []struct {
		start, end int
		want       string
	}{
		{0, 6, "012345"},
		{0, 3, "012"},
		{3, 6, "345"},
		{2, 4, "23"},
	}

	for _, tt := range tests {
		node := l.Slice(tt.start, tt.end)
		if node == nil {
			t.Errorf("Slice(%d, %d) returned nil", tt.start, tt.end)
			continue
		}
		// In case of full slice, it might return the same leaf or a new one,
		// but string value should match.
		// Our implementation: if start==0 && end==len, returns l.
		// If partial, returns new Leaf.

		val := ""
		if leaf, ok := node.(*Leaf); ok {
			val = leaf.val
		} else {
			t.Errorf("Slice returned node of type %T, want *Leaf", node)
		}

		if val != tt.want {
			t.Errorf("Slice(%d, %d) = %q, want %q", tt.start, tt.end, val, tt.want)
		}
	}
}

func TestTryMergeLeaves(t *testing.T) {
	// Case 1: Small strings should merge
	l1 := NewLeaf("hello")
	l2 := NewLeaf(" world")

	merged, ok := TryMergeLeaves(l1, l2)
	if !ok {
		t.Error("Expected merged Leaf, got false")
	}
	if merged.Len() != 11 {
		t.Errorf("Expected len 11, got %d", merged.Len())
	}
	if merged.String() != "hello world" {
		t.Errorf("Expected 'hello world', got %q", merged.String())
	}

	// Case 2: One is not a leaf
	c := NewConcat(l1, l2)
	_, ok = TryMergeLeaves(l1, c)
	if ok {
		t.Error("Expected false when merging Leaf with Concat")
	}

	// Case 3: Large strings should NOT merge (if exceeds MaxLeafMergeSize)
	// MaxLeafMergeSize is 256
	longStr := make([]byte, 200)
	for i := range longStr {
		longStr[i] = 'a'
	}
	s1 := string(longStr)

	longStr2 := make([]byte, 100) // 200+100 = 300 > 256
	for i := range longStr2 {
		longStr2[i] = 'b'
	}
	s2 := string(longStr2)

	large1 := NewLeaf(s1)
	large2 := NewLeaf(s2)

	_, ok = TryMergeLeaves(large1, large2)
	if ok {
		t.Error("Expected false (no merge) for large strings")
	}
}
