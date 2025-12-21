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

func TestConcat_Len(t *testing.T) {
	l1 := NewLeaf("hello")
	l2 := NewLeaf(" world")
	c := NewConcat(l1, l2)

	if got := c.Len(); got != 11 {
		t.Errorf("Len() = %d, want 11", got)
	}
}

func TestConcat_Depth(t *testing.T) {
	l1 := NewLeaf("a")
	l2 := NewLeaf("b")
	// l1, l2 depth = 0
	// c depth = 1 + max(0,0) = 1
	c := NewConcat(l1, l2)
	if d := c.Depth(); d != 1 {
		t.Errorf("Depth() = %d, want 1", d)
	}

	l3 := NewLeaf("c")
	// c2 depth = 1 + max(c.Depth(), l3.Depth()) = 1 + max(1, 0) = 2
	c2 := NewConcat(c, l3)
	if d := c2.Depth(); d != 2 {
		t.Errorf("Depth() = %d, want 2", d)
	}
}

func TestConcat_String(t *testing.T) {
	l1 := NewLeaf("foo")
	l2 := NewLeaf("bar")
	c := NewConcat(l1, l2)

	if got := c.String(); got != "foobar" {
		t.Errorf("String() = %q, want \"foobar\"", got)
	}
}

func TestConcat_ByteAt(t *testing.T) {
	l1 := NewLeaf("ABC")
	l2 := NewLeaf("DEF")
	c := NewConcat(l1, l2)

	tests := []struct {
		idx  int
		want byte
	}{
		{0, 'A'},
		{2, 'C'},
		{3, 'D'},
		{5, 'F'},
	}

	for _, tt := range tests {
		if got := c.ByteAt(tt.idx); got != tt.want {
			t.Errorf("ByteAt(%d) = %c, want %c", tt.idx, got, tt.want)
		}
	}
}

func TestConcat_Slice(t *testing.T) {
	// Setup: "01234" + "56789" = "0123456789"
	l1 := NewLeaf("01234")
	l2 := NewLeaf("56789")
	c := NewConcat(l1, l2) // length 10, split at 5

	tests := []struct {
		name       string
		start, end int
		want       string
	}{
		{"full slice", 0, 10, "0123456789"},
		{"left only", 0, 5, "01234"},
		{"left subset", 1, 4, "123"},
		{"right only", 5, 10, "56789"},
		{"right subset", 6, 9, "678"},
		{"spanning", 3, 7, "3456"},
		{"spanning start", 0, 7, "0123456"},
		{"spanning end", 3, 10, "3456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := c.Slice(tt.start, tt.end)
			if node.String() != tt.want {
				t.Errorf("Slice(%d, %d) = %q, want %q", tt.start, tt.end, node.String(), tt.want)
			}
		})
	}
}

func TestConcat_MarshalJSON(t *testing.T) {
	l1 := NewLeaf("hello")
	l2 := NewLeaf(" world")
	c := NewConcat(l1, l2)

	b, err := c.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	want := `"hello world"`
	if string(b) != want {
		t.Errorf("MarshalJSON = %s, want %s", string(b), want)
	}
}

func TestLeaf_Lines(t *testing.T) {
	tests := []struct {
		val  string
		want int
	}{
		{"", 0},
		{"hello", 0},
		{"hello\n", 1},
		{"\n", 1},
		{"a\nb\nc", 2},
		{"\n\n\n", 3},
	}
	for _, tt := range tests {
		l := NewLeaf(tt.val)
		if got := l.Lines(); got != tt.want {
			t.Errorf("Leaf(%q).Lines() = %d, want %d", tt.val, got, tt.want)
		}
	}
}

func TestConcat_Lines(t *testing.T) {
	// "one\n" (1) + "two\nthree" (1) = 2 lines
	l1 := NewLeaf("one\n")
	l2 := NewLeaf("two\nthree")
	c := NewConcat(l1, l2)

	if got := c.Lines(); got != 2 {
		t.Errorf("Concat lines = %d, want 2", got)
	}

	// Complex tree
	//       .
	//      / \
	//     .   \n (1)
	//    / \
	//   a   \n (1)
	l3 := NewLeaf("\n")
	l4 := NewLeaf("a")
	l5 := NewLeaf("\n")

	c2 := NewConcat(l4, l5) // 1 line
	c3 := NewConcat(c2, l3) // 1 + 1 = 2 lines

	if got := c3.Lines(); got != 2 {
		t.Errorf("Complex Concat lines = %d, want 2", got)
	}
}

func TestSlice_Lines(t *testing.T) {
	// "a\nb\nc\nd" -> 3 newlines
	// Indices:
	// 0: a
	// 1: \n
	// 2: b
	// 3: \n
	// 4: c
	// 5: \n
	// 6: d

	s := "a\nb\nc\nd"
	l := NewLeaf(s)

	// Slice "b\nc" -> indices 2 to 5.
	// 0123456
	// Slice(2, 5) -> "b\nc". Contains 1 newline.

	slice := l.Slice(2, 5)
	if slice.Lines() != 1 {
		t.Errorf("Slice lines = %d, want 1", slice.Lines())
	}

	// Slice across concat boundary
	// L1: "a\n"
	// L2: "b\n"
	// Concat(L1, L2) -> "a\nb\n"
	// Slice(1, 4) -> "\nb\n" -> 2 newlines

	c := NewConcat(NewLeaf("a\n"), NewLeaf("b\n"))
	// 0: a, 1: \n, 2: b, 3: \n
	// Slice 1 to 4: indices 1, 2, 3 -> "\nb\n"

	slice2 := c.Slice(1, 4)
	if slice2.String() != "\nb\n" {
		t.Fatalf("Slice content mismatch: %q", slice2.String())
	}
	if slice2.Lines() != 2 {
		t.Errorf("Concat Slice lines = %d, want 2", slice2.Lines())
	}
}

func TestTryMergeLeaves_Lines(t *testing.T) {
	l1 := NewLeaf("a\n") // 1 line
	l2 := NewLeaf("b\n") // 1 line

	merged, ok := TryMergeLeaves(l1, l2)
	if !ok {
		t.Fatal("Expected merge")
	}

	if merged.Lines() != 2 {
		t.Errorf("Merged lines = %d, want 2", merged.Lines())
	}
}
