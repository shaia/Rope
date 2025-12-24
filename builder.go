package rope

import (
	"fmt"

	"github.com/shaia/rope/internal/core"
)

// Balancer defines a strategy for combining two Rope nodes into a single Node.
// Implementations can control the structure of the resulting tree, enabling
// different balancing characteristics (e.g., AVL trees for general use, Fibonacci for append-heavy workloads).
type Balancer interface {
	// Join combines the left and right nodes into a new valid Rope Node.
	// It is responsible for maintaining any invariants associated with the balancing strategy
	// (e.g., height diffs in AVL). The resulting Node must contain the concatenation of
	// left's content followed by right's content.
	Join(left, right Node) Node
}

// Builder allows constructing and modifying Ropes with a specific configuration,
// such as a custom balancing strategy.
type Builder struct {
	balancer Balancer
}

// BuilderOption is a function that configures a Builder.
type BuilderOption func(*Builder)

// WithBalancer sets the balancing strategy for the Builder.
func WithBalancer(b Balancer) BuilderOption {
	return func(builder *Builder) {
		builder.balancer = b
	}
}

// NewBuilder creates a new Rope Builder with the given options.
// By default, it uses an AVL balancing strategy.
func NewBuilder(opts ...BuilderOption) *Builder {
	b := &Builder{
		balancer: core.NewAVLBalancer(),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Join combines two nodes using the configured balancing strategy.
func (b *Builder) Join(left, right Node) Node {
	return b.balancer.Join(left, right)
}

// Insert inserts a string into the rope at the given index using the configured balancer.
// Note: Insert involves slicing and re-joining, so the balancer impacts the resulting structure.
func (b *Builder) Insert(n Node, i int, text string) Node {
	if i < 0 || i > n.Len() {
		panic(fmt.Sprintf("index out of bounds: i=%d, valid range=[0,%d]", i, n.Len()))
	}

	// Create a new leaf for the inserted text
	inserted := New(text)

	if i == 0 {
		return b.Join(inserted, n)
	}
	if i == n.Len() {
		return b.Join(n, inserted)
	}

	left, right := Split(n, i)
	return b.Join(b.Join(left, inserted), right)
}

// Delete removes a range of bytes from the rope using the configured balancer.
func (b *Builder) Delete(n Node, start, end int) Node {
	if start < 0 || end > n.Len() || start > end {
		panic(fmt.Sprintf("index out of bounds or invalid range: start=%d, end=%d, valid range=[0,%d]", start, end, n.Len()))
	}

	if start == 0 && end == n.Len() {
		return New("")
	}

	// Split at start to get the left part and the rest
	left, rightPart := Split(n, start)

	// The rest (rightPart) contains the range to be deleted [0, end-start) relative to itself.
	// We split it at end-start. The right side of *that* split is what we want to keep.
	_, right := Split(rightPart, end-start)

	return b.Join(left, right)
}
