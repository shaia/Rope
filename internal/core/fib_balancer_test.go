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
	// We verify a tree that satisfies Fib invariants (Len >= Fib(Depth+2))
	// but violates AVL invariants (|dL - dR| > 1).

	// Construct the tree:
	// Left: Depth 2, Len 3. Structure: ((a,b), c)
	// Right: Depth 0, Len 3. Structure: "def"
	// Total: Depth 3, Len 6.
	// Fib check: Depth 3 requires Len >= Fib(5)=5.
	// We have Len 6. So it should NOT rebalance.
	// AVL check: |2 - 0| = 2 > 1. It WOULD rebalance in AVL.

	l1 := NewLeaf("a")
	l2 := NewLeaf("b")
	l3 := NewLeaf("c")

	c1 := NewConcat(l1, l2)   // Depth 1, Len 2
	left := NewConcat(c1, l3) // Depth 2, Len 3

	right := NewLeaf("def") // Depth 0, Len 3

	b := NewFibonacciBalancer()

	// Act
	root := b.Join(left, right)

	// Assert
	// 1. Check Depth is 3 (1 + max(2,0))
	if root.Depth() != 3 {
		t.Errorf("Expected Depth 3, got %d", root.Depth())
	}

	// 2. Check Len
	if root.Len() != 6 {
		t.Errorf("Expected Len 6, got %d", root.Len())
	}

	// 3. Verify it is structurally just Concat(left, right)
	// If it rebalanced, the structure would change (likely rotation).
	cRoot, ok := root.(*Concat)
	if !ok {
		t.Fatalf("Expected Concat node")
	}

	if cRoot.Left != left {
		t.Error("Balancer rebalanced (Left child changed) but shouldn't have")
	}
	if cRoot.Right != right {
		t.Error("Balancer rebalanced (Right child changed) but shouldn't have")
	}
}

func TestFibBalancer_Join_Empty(t *testing.T) {
	b := NewFibonacciBalancer()
	l := NewLeaf("a")
	e := NewLeaf("")

	// Join(l, empty) -> l
	res := b.Join(l, e)
	if res != l {
		t.Errorf("Expected left node to be returned directly when joining with empty, got %v", res)
	}

	// Join(empty, l) -> l
	res2 := b.Join(e, l)
	if res2 != l {
		t.Errorf("Expected right node to be returned directly when joining with empty, got %v", res2)
	}
}
