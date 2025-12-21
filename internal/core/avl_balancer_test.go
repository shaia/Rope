package core

import (
	"testing"
)

// Helper to create a leaf node with a specific depth for testing.
// In reality, leaves always have depth 0, but we might need deeply nested structures
// to trigger rotations.
// For this test, we will trust NewConcat to verify depth calculations
// and focusing on testing the Join logic of the Balancer.

func TestAVLBalancer_Join_Basic(t *testing.T) {
	b := NewAVLBalancer()
	// Use large strings to avoid coalescing (MaxLeafMergeSize=256)
	long := make([]byte, 200)
	l1 := NewLeaf(string(long))
	l2 := NewLeaf(string(long))

	// Join two leaves
	n := b.Join(l1, l2)
	if n.Len() != 400 {
		t.Errorf("Expected length 400, got %d", n.Len())
	}
	if n.Depth() != 1 {
		t.Errorf("Expected depth 1 (1+max(0,0)), got %d", n.Depth())
	}
}

func TestAVLBalancer_Join_Balance(t *testing.T) {
	// To test balancing, we need to create a tree that would be unbalanced if simply concatenated.
	//
	// Construct a tree:
	//       N
	//      / \
	//     L1  L2
	//    / \
	//   L3 L4
	//
	// Depths: L=0.
	// Concat(L, L) -> depth 1.
	// Concat(Concat(L,L), L) -> depth 2.

	b := NewAVLBalancer()

	// Create leaves
	// Create leaves large enough to avoid coalescing
	makeStr := func(char byte) string {
		s := make([]byte, 150)
		for i := range s {
			s[i] = char
		}
		return string(s)
	}
	leaf := func(s string) *Leaf { return NewLeaf(makeStr(s[0])) }

	// Left heavy case requiring rotation
	// We want to join ( (A, B), C ) with D??
	// No, Join takes two nodes.
	// Let's try to join a Depth 2 node with a Depth 0 node.
	// L3 = ((A,B), C) -> Depth 2
	// R1 = D -> Depth 0
	// Difference is 2. Should balance.

	n2 := b.Join(leaf("A"), leaf("B")) // Depth 1
	// Wait, if AVL is working, adding C to (A,B) might just make it balanced if C is added to the right?
	// (A,B) + C:
	// Depth(AB)=1, Depth(C)=0. Diff=1. No rotation needed usually, just new root.
	// New Root has left=AB(d1), right=C(d0). Depth = 2.

	// Now Join (Depth 2) with (Depth 0) -> D
	// Left (d2), Right (d0). Diff = 2.
	// Should trigger rotation.

	// Let's force it.
	// Left tree:
	//      . (d2)
	//     / \
	//   (d1) C(d0)
	//   /  \
	//  A(0) B(0)
	//
	// Right tree: D(0)

	// Join(Left, Right)
	// dLeft = 2, dRight = 0.
	// Condition: dLeft > dRight + 1 (2 > 1) -> True.
	// check lC.Right (C) vs lC.Left (AB).
	// lC is the root of Left. lC.Left is (AB), lC.Right is C.
	// d(AB)=1, d(C)=0.
	// lC.Right is NOT > lC.Left.
	// Single rotation case.

	// Result should be:
	//      .
	//     / \
	//    AB  (CD) ?? No.
	//
	// NewConcat(lC.Left, NewConcat(lC.Right, r))
	// Root -> Left=(AB), Right=(C, D)
	// Depth(AB)=1. Depth(CD)=1.
	// New Root Depth = 2.
	// (Previously it would be 3 if we just stuck them together: ((AB)C)D )

	lLeft := b.Join(n2, leaf("C"))
	lRight := leaf("D")

	balanced := b.Join(lLeft, lRight)

	if balanced.Depth() != 2 {
		t.Errorf("Expected balanced depth 2, got %d", balanced.Depth())
	}

	// Verify Structure manually if possible or just trust depth?
	// Accessing the structure:
	root, ok := balanced.(*Concat)
	if !ok {
		t.Fatalf("Expected Concat node")
	}

	// Expect Left to be (A,B) and Right to be (C,D) for a perfectly balanced tree?
	// Or at least better than linear.
	// In the single rotation logic:
	// return NewConcat(lC.Left, NewConcat(lC.Right, r))
	// lC was ((A,B), C). lC.Left=(A,B), lC.Right=C. r=D.
	// Result: ((A,B), (C,D))

	rLeft, ok := root.Left.(*Concat)
	if !ok {
		t.Errorf("Root.Left should be Concat")
	}
	if rLeft.Left.(*Leaf).val != makeStr('A') {
		t.Errorf("Expected A")
	}

	rRight, ok := root.Right.(*Concat)
	if !ok {
		t.Errorf("Root.Right should be Concat")
	}
	if rRight.Left.(*Leaf).val != makeStr('C') {
		t.Errorf("Expected C")
	}
}

func TestAVLBalancer_Empty(t *testing.T) {
	b := NewAVLBalancer()
	l := NewLeaf("a")
	// Empty node?
	// We don't have a public Empty node constructor in core yet easily accessible?
	// But Join checks Len() == 0.

	// Construct a fake empty node?
	// Or use NewLeaf("")
	e := NewLeaf("")

	res := b.Join(l, e)
	if res != l {
		// Optimization: should return non-empty node directly?
		// Join implementation: if right.Len() == 0 return left
		t.Error("Expected left node return for empty right")
	}

	res2 := b.Join(e, l)
	if res2 != l {
		t.Error("Expected right node return for empty left")
	}
}
