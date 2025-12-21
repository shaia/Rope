package core

// fibs holds precomputed Fibonacci numbers to estimate minimum length for a given depth.
// fib[0]=0, fib[1]=1...
// For a rope of depth d, the minimum length should be roughly fib[d+2].
// If len < fib[d+2], the tree is too sparse/deep and needs rebalancing.
var fibs []int

func init() {
	// Precompute fib numbers up to the max value of int.
	// On 64-bit systems, this goes up to index 91 (fib[91] fits in int64).
	// On 32-bit systems, likely index 46.
	fibs = []int{0, 1}
	for i := 2; ; i++ {
		next := fibs[i-1] + fibs[i-2]
		// Check for overflow. Since we are adding positive numbers,
		// result < prev indicates wrap-around (overflow).
		if next < fibs[i-1] {
			break
		}
		fibs = append(fibs, next)
	}
}

// FibonacciBalancer implements a lazy balancing strategy.
// It allows trees to become deeper than AVL, rebalancing only when
// the depth becomes inefficient regarding the number of nodes (length).
type FibonacciBalancer struct{}

func NewFibonacciBalancer() *FibonacciBalancer {
	return &FibonacciBalancer{}
}

func (b *FibonacciBalancer) Join(left, right Node) Node {
	// Optimization: Handle empty nodes
	if left.Len() == 0 {
		return right
	}
	if right.Len() == 0 {
		return left
	}

	// Optimization: Coalesce small leaves
	if merged, ok := TryMergeLeaves(left, right); ok {
		return merged
	}

	// 1. Cheap join
	concat := NewConcat(left, right)

	// 2. Check invariant: Is the tree "too deep"?
	// Standard check: Is length >= fib[depth+2]?
	// If NOT, we are too deep (sparse).
	d := concat.Depth()
	l := concat.Len()

	// Safety check for array bounds
	if d+2 >= len(fibs) {
		// Extremely deep. Definitely rebalance.
		return b.rebalance(concat)
	}

	minLen := fibs[d+2]
	if l < minLen {
		return b.rebalance(concat)
	}

	return concat
}

// rebalance flattens the rope and rebuilds it as a perfectly balanced tree.
func (b *FibonacciBalancer) rebalance(n Node) Node {
	leaves := collectLeaves(n)
	return buildBalanced(leaves)
}

func collectLeaves(n Node) []*Leaf {
	var leaves []*Leaf
	var visit func(n Node)
	visit = func(n Node) {
		if l, ok := n.(*Leaf); ok {
			leaves = append(leaves, l)
			return
		}
		if c, ok := n.(*Concat); ok {
			visit(c.Left)
			visit(c.Right)
		}
	}
	visit(n)
	return leaves
}

// buildBalanced constructs a perfectly balanced tree from a slice of leaves.
func buildBalanced(leaves []*Leaf) Node {
	if len(leaves) == 0 {
		return NewLeaf("") // Should not happen given logic
	}
	if len(leaves) == 1 {
		return leaves[0]
	}

	mid := len(leaves) / 2
	left := buildBalanced(leaves[:mid])
	right := buildBalanced(leaves[mid:])
	return NewConcat(left, right)
}
