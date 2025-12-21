package core

import (
	"testing"
)

// Helper functions in this test file construct unbalanced trees using NewConcat
// to simulate scenarios that trigger AVL rotations.
// While leaves always have Depth 0, we can build deeper Concat structures
// to test the Join method's rebalancing logic.

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
		t.Errorf("Expected depth 1, got %d", n.Depth())
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
	} else {
		rl, ok := rLeft.Left.(*Leaf)
		if !ok {
			t.Errorf("Expected Leaf for A")
		} else if rl.val != makeStr('A') {
			t.Errorf("Expected A")
		}
	}

	rRight, ok := root.Right.(*Concat)
	if !ok {
		t.Errorf("Root.Right should be Concat")
	} else {
		rr, ok := rRight.Left.(*Leaf)
		if !ok {
			t.Errorf("Expected Leaf for C")
		} else if rr.val != makeStr('C') {
			t.Errorf("Expected C")
		}
	}
}

func TestAVLBalancer_Join_Balance_RightHeavy(t *testing.T) {
	// Construct Right heavy case requiring rotation / rebalancing.
	// Join(Left(d0), Right(d2)) -> Diff = 2.
	// Should rebalance to Depth 2.

	b := NewAVLBalancer()
	makeStr := func(char byte) string {
		s := make([]byte, 150)
		for i := range s {
			s[i] = char
		}
		return string(s)
	}
	leaf := func(s string) *Leaf { return NewLeaf(makeStr(s[0])) }

	// Right tree (Depth 2)
	//      .
	//     / \
	//    A   B  (d1)
	//         \
	//          C (d0 - wait, usually leaves are d0. Concat(A,B)->d1.)
	// Need Depth 2.
	// R = (A, (B, C)) -> Right(d2)

	// Inner (B, C) -> Depth 1
	inner := b.Join(leaf("B"), leaf("C"))
	// Outer A + Inner -> Depth 2
	right := b.Join(leaf("A"), inner)

	if right.Depth() != 2 {
		t.Fatalf("Setup: Right depth %d != 2", right.Depth())
	}

	left := leaf("L") // Depth 0

	// Act
	n := b.Join(left, right)

	// Assert
	if n.Depth() != 2 {
		t.Errorf("Expected depth 2, got %d", n.Depth())
	}

	// Verify it's balanced
	// Expected: (L, A), (B, C) -> Depth 2.
	// Or similar balanced structure.

	root, ok := n.(*Concat)
	if !ok {
		t.Fatalf("Expected Concat root")
	}

	// Left child should be (L, A)
	rLeft, ok := root.Left.(*Concat)
	if !ok {
		t.Errorf("Expected Concat Left child")
	} else {
		// Check children of Left
		ll, ok := rLeft.Left.(*Leaf)
		if !ok || ll.val != makeStr('L') {
			t.Error("Left.Left should be L")
		}
		la, ok := rLeft.Right.(*Leaf)
		if !ok || la.val != makeStr('A') {
			t.Error("Left.Right should be A")
		}
	}

	// Right child should be (B, C) - actually "inner"
	rRight, ok := root.Right.(*Concat)
	if !ok {
		t.Errorf("Expected Concat Right child")
	} else {
		// Check children of Right
		rb, ok := rRight.Left.(*Leaf)
		if !ok || rb.val != makeStr('B') {
			t.Error("Right.Left should be B")
		}
		rc, ok := rRight.Right.(*Leaf)
		if !ok || rc.val != makeStr('C') {
			t.Error("Right.Right should be C")
		}
	}
}

func TestAVLBalancer_DoubleRotation_LR(t *testing.T) {
	// Construct Left-Right heavy tree.
	// Left heavy, but Inner child (Right of Left) is deeper.
	// Join(Left, Right)
	// Left: Concat(LL, LR) where LR > LL
	// Right: Leaf

	b := NewAVLBalancer()
	makeStr := func(char byte) string {
		s := make([]byte, 150)
		for i := range s {
			s[i] = char
		}
		return string(s)
	}
	leaf := func(s string) *Leaf { return NewLeaf(makeStr(s[0])) }

	// Build Inner Heavy Child LR (Depth 1)
	lr := NewConcat(leaf("LRL"), leaf("LRR")) // Depth 1

	// Build Left (Depth 2), but LL is Depth 0
	ll := leaf("LL")
	left := NewConcat(ll, lr) // Left(d2). Left.Right(d1) > Left.Left(d0).

	// Right is Leaf (Depth 0)
	right := leaf("R")

	// Pre-condition verification
	if left.Depth() != 2 {
		t.Fatalf("Setup error: Left depth %d != 2", left.Depth())
	}
	// left.Left depth 0, left.Right depth 1.

	// Act
	// Join(Left, Right)
	// dLeft(2) > dRight(0) + 1 -> Rebalance
	// Check Left.Right(d1) > Left.Left(d0) -> TRUE -> Double Rotation
	n := b.Join(left, right)

	// Assert
	// Should be balanced to Depth 2
	if n.Depth() != 2 {
		t.Errorf("Expected balanced depth 2, got %d", n.Depth())
	}

	// Verify Structure:
	// LR Rotation: NewConcat(NewConcat(l.Left, lr.Left), NewConcat(lr.Right, r))
	// Root.TopLeft = (LL, LRL)
	// Root.TopRight = (LRR, R)

	root, ok := n.(*Concat)
	if !ok {
		t.Fatalf("Expected Concat root")
	}

	// Check Top Left
	topLeft, ok := root.Left.(*Concat)
	if !ok {
		t.Error("TopLeft should be Concat")
	} else {
		// LL, LRL
		ll, ok := topLeft.Left.(*Leaf)
		if !ok {
			t.Error("Expected Leaf for LL")
		} else if ll.val != makeStr('L') { // "LL" starts with L
			t.Error("Expected LL")
		}

		lrl, ok := topLeft.Right.(*Leaf)
		if !ok {
			t.Error("Expected Leaf for LRL")
		} else if lrl.val != makeStr('L') { // "LRL" starts with L
			t.Error("Expected LRL")
		}
	}

	// Check Top Right
	topRight, ok := root.Right.(*Concat)
	if !ok {
		t.Error("TopRight should be Concat")
	} else {
		// LRR, R
		lrr, ok := topRight.Left.(*Leaf)
		if !ok {
			t.Error("Expected Leaf for LRR")
		} else if lrr.val != makeStr('L') { // "LRR" starts with L
			t.Error("Expected LRR")
		}

		r, ok := topRight.Right.(*Leaf)
		if !ok {
			t.Error("Expected Leaf for R")
		} else if r.val != makeStr('R') { // "R" starts with R
			t.Error("Expected R")
		}
	}
}

func TestAVLBalancer_DoubleRotation_RL(t *testing.T) {
	// Construct Right-Left heavy tree.
	// Right heavy, but Inner child (Left of Right) is deeper.
	// Join(Left, Right)
	// Left: Leaf (Depth 0)
	// Right: Concat(RL, RR) where RL > RR

	b := NewAVLBalancer()
	makeStr := func(char byte) string {
		s := make([]byte, 150)
		for i := range s {
			s[i] = char
		}
		return string(s)
	}
	leaf := func(s string) *Leaf { return NewLeaf(makeStr(s[0])) }

	// Build Inner Heavy Child RL (Depth 1)
	rl := NewConcat(leaf("RLL"), leaf("RLR")) // Depth 1

	// Build Right (Depth 2), RR is Depth 0
	rr := leaf("RR")
	right := NewConcat(rl, rr) // Right(d2). Right.Left(d1) > Right.Right(d0)

	// Left is Leaf (Depth 0)
	left := leaf("L")

	// Pre-condition
	if right.Depth() != 2 {
		t.Fatalf("Setup error: Right depth %d != 2", right.Depth())
	}

	// Act
	// Join(Left, Right)
	// dRight(2) > dLeft(0) + 1 -> Rebalance
	// Check Right.Left(d1) > Right.Right(d0) -> TRUE -> Double Rotation
	n := b.Join(left, right)

	// Assert
	if n.Depth() != 2 {
		t.Errorf("Expected balanced depth 2, got %d", n.Depth())
	}

	// Verify Structure
	// RL Rotation: NewConcat(NewConcat(l, rl.Left), NewConcat(rl.Right, rr))
	// Root.TopLeft = (L, RLL)
	// Root.TopRight = (RLR, RR)

	root, ok := n.(*Concat)
	if !ok {
		t.Fatalf("Expected Concat root")
	}

	// Check Top Left
	topLeft, ok := root.Left.(*Concat)
	if !ok {
		t.Error("TopLeft should be Concat")
	} else {
		l, ok := topLeft.Left.(*Leaf)
		if !ok {
			t.Error("Expected Leaf for L")
		} else if l.val != makeStr('L') {
			t.Error("Expected L")
		}

		rll, ok := topLeft.Right.(*Leaf)
		if !ok {
			t.Error("Expected Leaf for RLL")
		} else if rll.val != makeStr('R') { // RLL
			t.Error("Expected RLL")
		}
	}

	// Check Top Right
	topRight, ok := root.Right.(*Concat)
	if !ok {
		t.Error("TopRight should be Concat")
	} else {
		rlr, ok := topRight.Left.(*Leaf)
		if !ok {
			t.Error("Expected Leaf for RLR")
		} else if rlr.val != makeStr('R') { // RLR
			t.Error("Expected RLR")
		}

		rr, ok := topRight.Right.(*Leaf)
		if !ok {
			t.Error("Expected Leaf for RR")
		} else if rr.val != makeStr('R') { // RR
			t.Error("Expected RR")
		}
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
