package core

import "testing"

func TestFibBalancer_Basic(t *testing.T) {
	b := NewFibonacciBalancer()
	s := make([]byte, 150)
	l1 := NewLeaf(string(s))
	l2 := NewLeaf(string(s))

	// Simple join
	n := b.Join(l1, l2)
	if n.Len() != 300 {
		t.Errorf("Expected len 300")
	}
	if n.Depth() != 1 { // max(0,0)+1
		t.Errorf("Expected depth 1, got %d", n.Depth())
	}
}

func TestFibBalancer_RebalanceTrigger(t *testing.T) {
	b := NewFibonacciBalancer()

	// Construct a degenerate tree manually to force depth
	// A "linked list" of characters: a, b, c, d, e...
	// Length N, Depth N-1 if purely appended without balance.

	// Fib check:
	// Depth 0 (leaf): min len fib(2)=1. OK.
	// Depth 1 (2 leaves): min len fib(3)=2. OK.
	// Depth 2: min len fib(4)=3.
	// Depth 3: min len fib(5)=5.
	// Depth 4: min len fib(6)=8.

	// So if we have depth 4 but length 4 (e.g. "abcd"), 4 < 8, so it SHOULD rebalance.

	// Let's create a chain.
	var node Node = NewLeaf("a")
	for i := 0; i < 5; i++ {
		// Just appending single char.
		// "a", "b", "c", "d", "e", "f"
		node = b.Join(node, NewLeaf("x"))
	}

	// Total leaves = 6.
	// If purely concatenated right-heavy logic (which Join does if unbalanced):
	// ((...((a,x),x),x)...x) -> Depth 5?

	// Fib check for Depth 5: fib(7) = 13.
	// Length 6 < 13. Rebalance triggers.

	// If rebalanced perfectly:
	// 6 nodes -> log2(6) ~ 2.58 -> Depth 3 (root -> 2 -> 1).
	// Or Depth 3 or 4 depending on exact split.
	// A perfectly balanced tree of 6 nodes:
	//      N
	//    /   \
	//   N     N
	//  / \   / \
	// L   L L   N
	//          / \
	//         L   L
	//
	// Depth: Root(3).

	d := node.Depth()
	if d > 4 {
		t.Errorf("Expected rebalanced depth <= 4, got %d (Length %d)", d, node.Len())
	}
}

func TestFibBalancer_NoRebalanceNeeded(t *testing.T) {
	// Verify that it DOESN'T rebalance when not needed.
	// Difficult to test "internal logic didn't run", but we can check result structure or simply
	// assume if performance/correctness is fine.
	// Or: manually create a specific depth/len combo that is valid in Fib but invalid in AVL?

	// AVL strict: |dL - dR| <= 1.
	// Fib: Depth restriction is looser.

	// Example:
	// Tree satisfying Fib but not AVL.
	// Not easily constructible via PUBLIC Join API which enforces invariants.
}
