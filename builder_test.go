package rope

import (
	"testing"
)

// MockBalancer for testing injection
type MockBalancer struct {
	Called bool
}

func (m *MockBalancer) Join(left, right Node) Node {
	m.Called = true
	// just basic concat behavior to satisfy interface return
	// assuming we can verify the call in other ways
	// or we can delegate to AVL if we had access, but for now just return left if right is empty etc.
	if left.Len() == 0 {
		return right
	}
	if right.Len() == 0 {
		return left
	}

	// Create a dummy node or similar?
	// Since we can't easily create a 'Concat' node from outside core without using core,
	// let's just piggyback on default behavior via a trick or just return 'left' to prove interception?
	// If we return 'left', the resulting rope is just 'left', which is wrong but proves we intercepted.
	return left
}

func TestBuilder_Injection(t *testing.T) {
	mock := &MockBalancer{}
	b := NewBuilder(WithBalancer(mock))

	l1 := New("a")
	l2 := New("b")

	res := b.Join(l1, l2)

	if !mock.Called {
		t.Error("Custom balancer Join was not called")
	}

	// Since our mock returns 'left', res should be 'l1'
	if res != l1 {
		t.Error("Expected execution of mock logic")
	}
}

func TestBuilder_Default(t *testing.T) {
	b := NewBuilder()
	l1 := New("a")
	l2 := New("b")

	res := b.Join(l1, l2)

	if res.Len() != 2 {
		t.Errorf("Expected length 2, got %d", res.Len())
	}

	// Should be balanced (depth 2 for 2-leaf tree? Depth 1 if leaves are 0?
	// core tests established depth 1 for simple join of leaves.
	// Coalescing is enabled: "a" + "b" -> "ab" (Leaf, Depth 0)
	if res.Depth() != 0 {
		t.Errorf("Expected depth 0 (coalesced), got %d", res.Depth())
	}
}

func TestBuilder_Fibonacci(t *testing.T) {
	b := NewBuilder(WithBalancer(NewFibonacciBalancer()))

	// Create a chain to verify it works
	var node Node = New("start")
	for i := 0; i < 100; i++ {
		node = b.Join(node, New("x"))
	}

	if node.Len() != 105 { // 5 chars "start" + 100 * 1 char "x"
		t.Errorf("Expected length 105, got %d", node.Len())
	}
}
