package core

// AVLBalancer implements the Balancer interface using AVL tree rotations.
type AVLBalancer struct{}

// NewAVLBalancer creates a new instance of an AVL balancer.
func NewAVLBalancer() *AVLBalancer {
	return &AVLBalancer{}
}

// Join combines two nodes using AVL balancing logic.
func (b *AVLBalancer) Join(left, right Node) Node {
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

	dLeft := left.Depth()
	dRight := right.Depth()

	// If left tree is significantly deeper
	if dLeft > dRight+1 {
		lC, ok := left.(*Concat)
		if !ok {
			// Should not happen if depths are correct for leaves vs concats
			// But strictly speaking, a Leaf has depth 1.
			return NewConcat(left, right)
		}

		// Recursively join the right child of left with the new right node
		// This might need rebalancing
		balancedRight := b.Join(lC.Right, right)
		return b.balance(lC.Left, balancedRight)
	}

	// If right tree is significantly deeper
	if dRight > dLeft+1 {
		rC, ok := right.(*Concat)
		if !ok {
			return NewConcat(left, right)
		}

		// Recursively join left with the left child of right
		balancedLeft := b.Join(left, rC.Left)
		return b.balance(balancedLeft, rC.Right)
	}

	// Logic for roughly equal depths: just concatenate
	return NewConcat(left, right)
}

// balance acts as the rotation mechanism, similar to the legacy balance.go logic
func (b *AVLBalancer) balance(l, r Node) Node {
	const diffThreshold = 1
	dDiff := l.Depth() - r.Depth()

	if dDiff > diffThreshold {
		lC, ok := l.(*Concat)
		if !ok {
			return NewConcat(l, r)
		}
		// Double rotation check: if left-right is deeper than left-left
		if lC.Right.Depth() > lC.Left.Depth() {
			lrC, ok2 := lC.Right.(*Concat)
			if ok2 {
				// Left-Right rotation
				return NewConcat(
					NewConcat(lC.Left, lrC.Left),
					NewConcat(lrC.Right, r),
				)
			}
		}
		// Single rotation (Right rotation)
		return NewConcat(lC.Left, NewConcat(lC.Right, r))
	} else if dDiff < -diffThreshold {
		rC, ok := r.(*Concat)
		if !ok {
			return NewConcat(l, r)
		}
		// Double rotation check: if right-left is deeper than right-right
		if rC.Left.Depth() > rC.Right.Depth() {
			rlC, ok2 := rC.Left.(*Concat)
			if ok2 {
				// Right-Left rotation
				return NewConcat(
					NewConcat(l, rlC.Left),
					NewConcat(rlC.Right, rC.Right),
				)
			}
		}
		// Single rotation (Left rotation)
		return NewConcat(NewConcat(l, rC.Left), rC.Right)
	}

	return NewConcat(l, r)
}
