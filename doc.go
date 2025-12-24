// Package rope implements a high-performance, concurrent, immutable Rope data structure (Rope).
//
// A Rope is a tree-based data structure used for storing strings. It is significantly
// more efficient than standard strings or byte slices for large texts that undergo
// frequent modification (insertions, deletions, concatenations).
//
// Features:
//   - Immutable: All operations return a new Node, sharing unchanged structure with the original.
//   - Concurrent: Safe for concurrent readers without locks. Thread-safe `RopeHandle` for atomic updates.
//   - Efficient: O(log N) for Insert, Delete, Split, and Concat operations.
//   - Search: Efficient Rabin-Karp substring search.
//   - Indexing: Line and Column tracking (0-indexed) with O(log N) lookups.
//   - Flexible: Pluggable balancing strategies (AVL, Fibonacci) and "Builder" pattern.
//   - IO-Friendly: Implements `io.Reader`, `io.WriterTo`, and `json.Marshaler`.
//
// This package targets high-performance text editing and processing scenarios.
package rope
