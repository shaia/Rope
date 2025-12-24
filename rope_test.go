package rope

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestRopeBasics(t *testing.T) {
	r1 := New("Hello")
	r2 := New(" World")
	r3 := ConcatNodes(r1, r2)

	if r3.String() != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", r3.String())
	}

	if r3.Len() != 11 {
		t.Errorf("Expected length 11, got %d", r3.Len())
	}
}

func TestSlice(t *testing.T) {
	r := New("Hello World")
	s1 := r.Slice(0, 5)
	if s1.String() != "Hello" {
		t.Errorf("Slice(0, 5) failed: got '%s'", s1.String())
	}

	s2 := r.Slice(6, 11)
	if s2.String() != "World" {
		t.Errorf("Slice(6, 11) failed: got '%s'", s2.String())
	}
}

func TestInsert(t *testing.T) {
	r := New("HelloWorld")
	r2 := Insert(r, 5, " ")
	if r2.String() != "Hello World" {
		t.Errorf("Insert failed: got '%s'", r2.String())
	}
}

func TestDelete(t *testing.T) {
	r := New("Hello World")
	r2 := Delete(r, 5, 6) // Delete space
	if r2.String() != "HelloWorld" {
		t.Errorf("Delete failed: got '%s'", r2.String())
	}
	// Test case requested by user: Delete middle range
	// "HelloWorld" (len 10). Split 2, 7.
	// Indices: 0 1 | 2 3 4 5 6 | 7 8 9
	// Chars:   H e | l l o W o | r l d
	// Deleting [2, 7) should remove "lloWo".
	// Expected result: "He" + "rld" = "Herld"
	r3 := New("HelloWorld")
	r4 := Delete(r3, 2, 7)
	if r4.String() != "Herld" {
		t.Errorf("Delete(2, 7) failed: expected 'Herld', got '%s'", r4.String())
	}
}

func TestComplexOps(t *testing.T) {
	// Construct a deeper tree
	r := New("")
	for _, word := range []string{"This", " ", "is", " ", "a", " ", "rope"} {
		r = ConcatNodes(r, New(word))
	}
	expected := "This is a rope"
	if r.String() != expected {
		t.Errorf("Complex build failed: got '%s'", r.String())
	}

	// Depth check - for small strings, it might be coalesced to 0 depth.
	// So we won't enforce depth > 0 here.
	// if r.Depth() == 0 && len(expected) > 0 {
	// 	 t.Errorf("Expected depth > 0 for multi-node rope")
	// }

	// Create a large rope to verify depth > 0
	word := "This is a rope "
	rLarge := New("")
	for i := 0; i < 20; i++ {
		rLarge = ConcatNodes(rLarge, New(word))
	}
	if rLarge.Depth() == 0 {
		t.Errorf("Expected depth > 0 for large rope (len %d)", rLarge.Len())
	}

	// Insert in middle
	r2 := Insert(r, 9, "n efficient")
	// "This is a rope" -> insert at 9 (after 'a') -> "This is an efficient rope"
	// 0123456789
	// This is a^
	expected2 := "This is an efficient rope"
	if r2.String() != expected2 {
		t.Errorf("Complex insert failed: expected '%s', got '%s'", expected2, r2.String())
	}

	// Original should be unchanged
	if r.String() != expected {
		t.Errorf("Immutability violation: original changed to '%s'", r.String())
	}
}

func TestEmptyOps(t *testing.T) {
	e1 := New("")
	e2 := New("")

	// Join empty
	j := Join(e1, e2)
	if j.Len() != 0 {
		t.Errorf("Join(empty, empty) len = %d", j.Len())
	}

	// Insert into empty
	i := Insert(e1, 0, "hello")
	if i.String() != "hello" {
		t.Errorf("Insert into empty failed: %s", i.String())
	}

	// Delete from empty (should be no-op or valid empty return if range 0,0)
	d := Delete(e1, 0, 0)
	if d.Len() != 0 {
		t.Errorf("Delete(empty) failed")
	}
}

func TestUnicode(t *testing.T) {
	// Rope is byte-based, but we deal with strings.
	// Ensure high-byte characters are preserved in order.
	s := "Hello 🌍 World"
	r := New(s)

	if r.Len() != len(s) {
		t.Errorf("Len() mismatch for unicode. Got %d, want %d", r.Len(), len(s))
	}

	if r.String() != s {
		t.Errorf("String() mismatch. Got %s, want %s", r.String(), s)
	}

	// Split inside the globe character?
	// The library doesn't prevent splitting bytes of a utf8 char,
	// but rebuilding strict should be correct.
	mid := 7 // "Hello " is 6 bytes. Index 7 splits inside the 4-byte 🌍 character.

	left, right := Split(r, mid)
	joined := Join(left, right)

	if joined.String() != s {
		t.Errorf("Split/Join unicode corruption. Got %s", joined.String())
	}
}

func TestEdgeCases(t *testing.T) {
	s := "test"
	r := New(s)

	// Split at 0
	l, right := Split(r, 0)
	if l.Len() != 0 || right.String() != s {
		t.Errorf("Split at 0 failed")
	}

	// Split at end
	left, r2 := Split(r, 4)
	if left.String() != s || r2.Len() != 0 {
		t.Errorf("Split at end failed")
	}

	// Delete all
	d := Delete(r, 0, 4)
	if d.Len() != 0 {
		t.Errorf("Delete all failed")
	}
}

func TestLargeBalance(t *testing.T) {
	// 64 nodes of length 1 -> balanced Depth should be roughly log2(64) ~ 6 or 7
	// Naive concat would be depth 64
	r := New("a")
	for i := 0; i < 63; i++ {
		r = Join(r, New("a"))
	}

	if r.Len() != 64 {
		t.Fatalf("Length mismatch: %d", r.Len())
	}

	// Allowing some slack for AVL implementation details, but 64 is way too high
	if r.Depth() > 10 {
		t.Errorf("Tree not balanced. Depth %d for 64 nodes", r.Depth())
	}
}

func TestRopeHandle_Basic(t *testing.T) {
	h := NewHandle(New("initial"))

	if h.Root().String() != "initial" {
		t.Errorf("Expected initial content")
	}

	h.Set(New("updated"))
	if h.Root().String() != "updated" {
		t.Errorf("Expected updated content")
	}

	h.Apply(func(n Node) Node {
		return Join(n, New("!"))
	})

	if h.Root().String() != "updated!" {
		t.Errorf("Expected applied content")
	}
}

func TestRopeHandle_Concurrency(t *testing.T) {
	h := NewHandle(New(""))

	// N writers appending "a"
	runners := 100
	writesPerRunner := 100

	var wg sync.WaitGroup
	wg.Add(runners)

	for i := 0; i < runners; i++ {
		go func() {
			defer wg.Done()
			for k := 0; k < writesPerRunner; k++ {
				h.Apply(func(n Node) Node {
					return Join(n, New("a"))
				})
			}
		}()
	}

	// Concurrent readers
	readerStop := make(chan struct{})
	var readWg sync.WaitGroup
	readWg.Add(1)
	go func() {
		defer readWg.Done()
		for {
			select {
			case <-readerStop:
				return
			default:
				r := h.Root()
				// access len/depth safely
				_ = r.Len()
				_ = r.String()
			}
		}
	}()

	wg.Wait()
	close(readerStop)
	readWg.Wait()

	final := h.Root()
	expectedLen := runners * writesPerRunner // 100 * 10 = 1000
	if final.Len() != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, final.Len())
	}

	// Verify content is all 'a's
	str := final.String()
	for i, r := range str {
		if r != 'a' {
			t.Errorf("Expected 'a' at %d", i)
			break
		}
	}

	t.Logf("Final Rope Depth: %d", final.Depth())
}

func TestRopeHandle_Snapshot(t *testing.T) {
	h := NewHandle(New("v1"))
	snap1 := h.Snapshot()

	if snap1.String() != "v1" {
		t.Errorf("Expected v1")
	}

	// Update handle
	h.Set(New("v2"))

	// Snapshot should be unchanged
	if snap1.String() != "v1" {
		t.Errorf("Snapshot mutated! Expected v1, got %s", snap1.String())
	}

	// New snapshot
	snap2 := h.Snapshot()
	if snap2.String() != "v2" {
		t.Errorf("Expected v2")
	}
}

func TestParForEach(t *testing.T) {
	// Build a rope with multiple chunks
	// "chunk0", "chunk1", ... "chunk9"
	chunks := []string{}
	r := New("")
	for i := 0; i < 100; i++ {
		s := fmt.Sprintf("chunk%d ", i)
		chunks = append(chunks, s)
		r = ConcatNodes(r, New(s))
	}

	expectedFull := ""
	for _, c := range chunks {
		expectedFull += c
	}

	// 1. Single worker (sequential effectively)
	{
		var mu sync.Mutex
		collected := ""
		ParForEach(r, 1, func(s string) {
			mu.Lock()
			collected += s
			mu.Unlock()
		})
		// Order should be preserved if traversal is recursive DFS and single worker?
		// Actually, even with 1 worker, we use a channel. Channel order is preserved.
		// Traversal pushes in order. Worker pulls in order.
		if collected != expectedFull {
			t.Errorf("Single worker ParForEach mismatch. Len %d vs %d", len(collected), len(expectedFull))
		}
	}

	// 2. Multiple workers (parallel)
	{
		var mu sync.Mutex
		collectedParts := []string{}
		ParForEach(r, 10, func(s string) {
			mu.Lock()
			collectedParts = append(collectedParts, s)
			mu.Unlock()
		})

		// Verify we got all parts. Order is NOT guaranteed.
		// Sort both and compare? Or just reconstruct string if unique?
		// "chunkX " strings are unique for different X (mostly).

		// Let's verify total length and content match.
		totalLen := 0
		for _, p := range collectedParts {
			totalLen += len(p)
		}

		// Reconstruct original chunks map to verify presence
		// NOTE: Since TryMergeLeaves is active (MaxLeafMergeSize=256), many small "chunkX "
		// strings will be coalesced into larger leaves.
		// Thus, we cannot expect `collectedParts` to match `chunks` exactly.
		// Instead, we verify that `collectedParts` allows reconstructing the full string,
		// OR simply that every char count matches.

		// For simplicity/correctness, let's verify if we concat all parts, do we have the same set of characters?
		// Since order is lost, simple string equality won't work.
		// Let's rely on total length (already checked) and verify that all original chunks are present in the *concatenation* of collected parts if we were to treat them as a bag of substrings?
		// Better: check that every collected part is a valid substring of expectedFull?
		// (Assuming no duplicate content allows ambiguity). E.g. "chunk1 chunk2" is a valid substring.

		for _, p := range collectedParts {
			if !strings.Contains(expectedFull, p) {
				t.Errorf("Collected part %q is seemingly not part of original content", p)
			}
		}

		// Also verify total length again just to be sure we didn't miss/add anything in size.
		if totalLen != len(expectedFull) {
			t.Errorf("Total length mismatch. Got %d, want %d", totalLen, len(expectedFull))
		}
	}
}

func TestParForEach_Large(t *testing.T) {
	// Simple smoke test for larger tree
	r := New("init")
	for i := 0; i < 1000; i++ {
		r = ConcatNodes(r, New("a"))
	}

	count := 0
	var mu sync.Mutex

	ParForEach(r, 8, func(s string) {
		mu.Lock()
		count += len(s)
		mu.Unlock()
	})

	if count != 1004 { // "init" (4) + 1000 * "a" (1)
		t.Errorf("Large parallel mismatch. Got len %d", count)
	}
}
