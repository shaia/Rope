package tests

import (
	"math/rand"
	"testing"

	"github.com/shaia/rope"
)

// BenchmarkInsertSequential benchmarks appending to the end.
func BenchmarkInsertSequential(b *testing.B) {
	initial := rope.New("Initial Content")
	text := "0123456789"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Just repeated insertions, creating a very deep/heavy structure if not balanced.
		// Note: We use the result to prevent compiler optimizations,
		// but we don't chain it deeply in this loop variable sense locally
		// unless we want to test O(N) memory.
		// Testing "Edit at random position" is more realistic usually.
		// But let's test building a huge rope.
		_ = rope.Insert(initial, 5, text)
	}
}

// BenchmarkBuildLarge builds a large rope by continuously appending.
// This stress-tests the rebalancing logic.
func BenchmarkBuildLarge(b *testing.B) {
	text := "SmallChunk"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := rope.New("")
		for j := 0; j < 1000; j++ {
			r = rope.Join(r, rope.New(text))
		}
	}
}

// BenchmarkRandomEdits simulates a user editing a file.
func BenchmarkRandomEdits(b *testing.B) {
	// Setup a large initial rope
	r := rope.New(generateRandomString(10000))

	// Pre-generate edits to avoid benchmarking rand() too much
	indices := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		indices[i] = rand.Intn(10000)
	}

	insertText := "x"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r = rope.Insert(r, indices[i], insertText)
	}
}

// BenchmarkSlice simulates scrolling/viewing random parts.
func BenchmarkSlice(b *testing.B) {
	r := rope.New(generateRandomString(100000))
	maxIdx := 99000

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := i % maxIdx
		_ = r.Slice(start, start+100)
	}
}

// BenchmarkConcatLarge benchmarks joining two large ropes.
func BenchmarkConcatLarge(b *testing.B) {
	s1 := generateRandomString(50000)
	s2 := generateRandomString(50000)
	r1 := rope.New(s1)
	r2 := rope.New(s2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rope.Join(r1, r2)
	}
}

// BenchmarkSplit benchmarks splitting a large rope.
func BenchmarkSplit(b *testing.B) {
	r := rope.New(generateRandomString(100000))
	splitIdx := 50000

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = rope.Split(r, splitIdx)
	}
}

// BenchmarkDelete benchmarks deleting a range from a rope.
func BenchmarkDelete(b *testing.B) {
	r := rope.New(generateRandomString(10000))
	start := 4000
	length := 2000

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rope.Delete(r, start, start+length)
	}
}

// BenchmarkString benchmarks converting the rope back to a string.
// This is allocation heavy.
func BenchmarkString(b *testing.B) {
	r := rope.New(generateRandomString(10000))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.String()
	}
}

// BenchmarkSearch benchmarks the Rabin-Karp substring search.
func BenchmarkSearch(b *testing.B) {
	// Create a large text with the pattern at the end
	text := generateRandomString(100000) + "TARGET"
	r := rope.New(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rope.Index(r, "TARGET")
	}
}

// BenchmarkAppend benchmarks appending to the end (special case of Insert).
func BenchmarkAppend(b *testing.B) {
	initial := rope.New("Initial Content")
	toAppend := "0123456789"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// New rope every time to avoid growing indefinitely in memory (which tests GC more than logic)
		// Or should we grow it?
		// Existing BenchmarkInsertSequential grows it.
		// Let's test the op cost on a static baserope
		_ = rope.Insert(initial, initial.Len(), toAppend)
	}
}

// BenchmarkPrepend benchmarks inserting at the beginning.
func BenchmarkPrepend(b *testing.B) {
	initial := rope.New("Initial Content")
	toPrepend := "0123456789"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rope.Insert(initial, 0, toPrepend)
	}
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
