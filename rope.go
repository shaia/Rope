package rope

import (
	"encoding/json"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/shaia/rope/internal/core"
)

// Node represents a node in the immutable rope.
type Node = core.Node

// defaultBuilder is the singleton builder used for package-level functions.
// It uses the standard AVL balancing strategy.
var defaultBuilder = NewBuilder()

// New creates a new Rope from a string.
// For short strings, it returns a Leaf.
// For long strings, it might split it.
func New(s string) Node {
	return core.NewLeaf(s)
}

// NewFibonacciBalancer creates a new balancer that uses a lazy rebalancing strategy.
// It is optimized for append-heavy workloads.
func NewFibonacciBalancer() Balancer {
	return core.NewFibonacciBalancer()
}

// Join concatenates two Ropes.
// It uses the default AVL balancing strategy.
func Join(a, b Node) Node {
	return defaultBuilder.Join(a, b)
}

// ConcatNodes is the public API for joining nodes.
func ConcatNodes(a, b Node) Node {
	return Join(a, b)
}

// Split cuts the rope at index i, returning two new Ropes.
func Split(n Node, i int) (Node, Node) {
	if i == 0 {
		return New(""), n
	}
	if i == n.Len() {
		return n, New("")
	}
	return n.Slice(0, i), n.Slice(i, n.Len())
}

// Insert inserts the string text into the Rope n at the 0-based index i.
// It returns a new Rope (Node) representing the modified string.
// The original Rope n is unmodified.
// Panics if i is out of bounds (i < 0 or i > n.Len()).
func Insert(n Node, i int, text string) Node {
	return defaultBuilder.Insert(n, i, text)
}

// Delete removes the range of bytes [start, end) from the Rope n.
// It returns a new Rope (Node) representing the modified string.
// The original Rope n is unmodified.
// Panics if start or end are out of bounds.
func Delete(n Node, start, end int) Node {
	return defaultBuilder.Delete(n, start, end)
}

// RopeHandle provides a thread-safe wrapper around a Rope.
// It allows for lock-free reads (snapshots) and serialized writes.
type RopeHandle struct {
	// value holds the *NodeContainer
	value atomic.Value
	mu    sync.Mutex
}

// container is a helper to store the interface in atomic.Value
type container struct {
	root Node
}

// NewHandle creates a new thread-safe handle for a Rope.
func NewHandle(initial Node) *RopeHandle {
	h := &RopeHandle{}
	h.value.Store(&container{root: initial})
	return h
}

// Root returns the current snapshot of the Rope.
// This operation is O(1), wait-free, and thread-safe.
func (h *RopeHandle) Root() Node {
	return h.value.Load().(*container).root
}

// Snapshot is an alias for Root, emphasizing the persistent nature of the data structure.
// It returns a point-in-time view of the Rope that will not change even if the Handle is updated.
func (h *RopeHandle) Snapshot() Node {
	return h.Root()
}

// Set updates the root of the Rope safely.
func (h *RopeHandle) Set(n Node) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.value.Store(&container{root: n})
}

// Apply atomically applies a modification function to the rope.
// function `fn` receives the current root and returns the new root.
// Writes are serialized using a mutex to prevent race conditions during the read-modify-write cycle.
func (h *RopeHandle) Apply(fn func(Node) Node) Node {
	h.mu.Lock()
	defer h.mu.Unlock()

	current := h.value.Load().(*container).root
	newRoot := fn(current)
	h.value.Store(&container{root: newRoot})
	return newRoot
}

// MarshalJSON marshals the rope's content as a JSON string.
func (h *RopeHandle) MarshalJSON() ([]byte, error) {
	// Snapshot the current state
	snap := h.Snapshot()
	// Check if the root node is nil (e.g. NewHandle(nil) was called).
	if snap == nil {
		return json.Marshal("")
	}
	// Warning: Materializes full string
	return json.Marshal(snap.String())
}

// UnmarshalJSON unmarshals a JSON string into the rope handle, REPLACING its content.
func (h *RopeHandle) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	h.Set(New(s))
	return nil
}

// ParForEach iterates over the leaves of the Rope in parallel.
// It uses a specified number of worker goroutines to process string chunks.
// The `fn` callback is executed for each leaf's string content.
// Since execution is concurrent, the order of processing is NOT guaranteed.
func ParForEach(n Node, workers int, fn func(string)) {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	// Channel to feed leaves to workers
	jobs := make(chan string, workers*2)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for s := range jobs {
				fn(s)
			}
		}()
	}

	// Traverse the tree and send leaves to jobs channel
	traverse(n, jobs)
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
}

func traverse(n Node, jobs chan<- string) {
	if n == nil || n.Len() == 0 {
		return
	}
	n.EachLeaf(func(s string) {
		jobs <- s
	})
}
