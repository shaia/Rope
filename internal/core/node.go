package core

import (
	"encoding/json"
	"strings"
)

// Node represents a node in the immutable rope.
// It matches the public interface but is defined here to allow internal types to refer to it.
type Node interface {
	Len() int
	ByteAt(idx int) byte
	Slice(start, end int) Node
	String() string
	Depth() int
	Lines() int
}

// --------------------------------------------------------
// Leaf Implementation
// --------------------------------------------------------

// Leaf represents a leaf node containing a string segment.
type Leaf struct {
	val   string
	lines int
}

// NewLeaf creates a new leaf node.
func NewLeaf(s string) *Leaf {
	return &Leaf{
		val:   s,
		lines: strings.Count(s, "\n"),
	}
}

func (l *Leaf) Len() int {
	return len(l.val)
}

func (l *Leaf) ByteAt(idx int) byte {
	return l.val[idx]
}

func (l *Leaf) String() string {
	return l.val
}

func (l *Leaf) Depth() int {
	return 0
}

func (l *Leaf) Lines() int {
	return l.lines
}

// MarshalJSON returns the JSON encoding of the leaf's string value.
func (l *Leaf) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.val)
}

func (l *Leaf) Slice(start, end int) Node {
	if start == 0 && end == len(l.val) {
		return l
	}
	// Bounds check should be handled, or standard Go slice panic will occur
	sub := l.val[start:end]
	return &Leaf{
		val:   sub,
		lines: strings.Count(sub, "\n"),
	}
}

// MaxLeafMergeSize represents the maximum size in bytes for a leaf node created via merging.
// Leaves smaller than this threshold will be coalesced to avoid creating unnecessary Concat nodes.
const MaxLeafMergeSize = 256

// TryMergeLeaves attempts to merge two nodes into a single leaf if they are both leaves
// and their combined length is small enough.
// Returns the merged node and true if successful, or nil and false otherwise.
func TryMergeLeaves(left, right Node) (Node, bool) {
	// Fast check for lengths first
	if left.Len()+right.Len() > MaxLeafMergeSize {
		return nil, false
	}

	lLeaf, ok1 := left.(*Leaf)
	rLeaf, ok2 := right.(*Leaf)

	if ok1 && ok2 {
		mergedVal := lLeaf.val + rLeaf.val
		return &Leaf{
			val:   mergedVal,
			lines: lLeaf.lines + rLeaf.lines, // Optimization: sum counts instead of rescanning
		}, true
	}

	return nil, false
}

// --------------------------------------------------------
// Concat Implementation
// --------------------------------------------------------

// Concat represents an internal node concatenating two children.
type Concat struct {
	Left, Right Node
	length      int
	depth       int
	lines       int
}

// NewConcat creates a new concat node.
// It automatically balances if necessary (TODO).
func NewConcat(left, right Node) *Concat {
	return &Concat{
		Left:   left,
		Right:  right,
		length: left.Len() + right.Len(),
		depth:  1 + max(left.Depth(), right.Depth()),
		lines:  left.Lines() + right.Lines(),
	}
}

// JoinNodes combines two nodes into a single Node using the default AVL strategy.
// This function is kept for backward compatibility and internal convenience.
func JoinNodes(left, right Node) Node {
	balancer := NewAVLBalancer()
	return balancer.Join(left, right)
}

func (c *Concat) Len() int {
	return c.length
}

func (c *Concat) Depth() int {
	return c.depth
}

func (c *Concat) Lines() int {
	return c.lines
}

func (c *Concat) String() string {
	return c.Left.String() + c.Right.String()
}

func (c *Concat) ByteAt(idx int) byte {
	leftLen := c.Left.Len()
	if idx < leftLen {
		return c.Left.ByteAt(idx)
	}
	return c.Right.ByteAt(idx - leftLen)
}

// MarshalJSON returns the JSON encoding of the rope's full string content.
// Note: This may be expensive as it materializes the entire rope into a string.
func (c *Concat) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Concat) Slice(start, end int) Node {
	if start == 0 && end == c.length {
		return c
	}
	leftLen := c.Left.Len()

	// Case 1: Slice is entirely in the left child
	if end <= leftLen {
		return c.Left.Slice(start, end)
	}

	// Case 2: Slice is entirely in the right child
	if start >= leftLen {
		return c.Right.Slice(start-leftLen, end-leftLen)
	}

	// Case 3: Slice spans both children
	return JoinNodes(
		c.Left.Slice(start, leftLen),
		c.Right.Slice(0, end-leftLen),
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
