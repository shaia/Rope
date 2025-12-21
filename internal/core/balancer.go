package core

// Balancer defines the interface for combining two nodes into a balanced tree.
type Balancer interface {
	// Join combines two nodes (left and right) into a single Node.
	// It is responsible for maintaining the structural invariants of the implementation (e.g., AVL balance).
	Join(left, right Node) Node
}
