package rope

import (
	"strings"

	"github.com/shaia/rope/internal/core"
)

// RowColToOffset converts a 0-based row and column index to a byte offset.
// Returns -1 if the row/col is out of bounds.
// Note: This matches "visual" lines based on '\n'.
func RowColToOffset(n Node, row, col int) int {
	if n == nil {
		return -1
	}
	if row < 0 || col < 0 {
		return -1
	}
	// Optimization: If row > total lines, fast fail
	if row > n.Lines() {
		return -1
	}
	return rowColToOffsetRec(n, row, col)
}

func rowColToOffsetRec(n Node, row, col int) int {
	// Base case: Leaf
	if l, ok := n.(*core.Leaf); ok {
		// We need to find the offset of the Nth newline in this string?
		// This is actually O(LeafLen), which is small.
		// If row > l.Lines(), then it's out of bounds of this leaf,
		// but the parent logic should have directed us correctly.
		// Exception: searching for end of file position?

		s := l.String()
		currentLine := 0
		offset := 0
		for i, r := range s {
			if currentLine == row {
				// check column
				if i-offset == col {
					return i
				}
			}
			if r == '\n' {
				currentLine++
				offset = i + 1
				if currentLine > row {
					// Passed the target row
					return -1
				}
			}
		}
		// If we are on the correct row and col is exactly at the end (for last line without newline)
		if currentLine == row {
			if len(s)-offset == col {
				return len(s)
			}
		}

		return -1
	}

	// Recursive case: Concat
	if c, ok := n.(*core.Concat); ok {
		leftLines := c.Left.Lines()
		if row < leftLines {
			// Target is in left child
			return rowColToOffsetRec(c.Left, row, col)
		} else if row == leftLines {
			// Target *might* be in left child if left child ends with NO newline and we are continuing that line?
			// The contract of "Lines" is count of '\n'.
			// Logic:
			// If Left has K newlines, it contains lines 0..K-1 fully.
			// Line K starts in Left (after last \n) and continues to Right?
			// Actually, "Lines()" returns count of '\n'.
			// Example: "A\nB" -> Lines=1. Row 0 is "A", Row 1 is "B".
			// If row < leftLines: definitely in left.
			// If row == leftLines: It starts in Left after the last \n of Left?
			// No, standard convention:
			// A line belongs to the node containing the characters.
			// A split line spans both.
			// This makes "Lines()" slightly ambiguous for indexing.
			// Usually we need `Lines()` to mean "number of fully completed lines".

			// Let's assume standard behavior:
			// If target row < leftLines: strictly inside left's fully completed lines.
			// If target row == leftLines: It's the "partial" line at the end of Left, continuing into Right (starting at 0 in Right).
			// WAIT. If row == leftLines, we need to check if we fit in the remaining part of Left?
			// Complexity of split lines is real.
			// SIMPLIFICATION:
			// Let's delegate:
			// 1. Calculate offset in Left of the start of row `leftLines`.
			//    But getting that is hard.

			// Better Strategy:
			// Maintain global context? No.
			// Let's implement simpler: OffsetToRowCol is cleaner.
			// For RowColToOffset, maybe we iterate?

			// Let's stick to the plan: traverse.
			// If row < c.Left.Lines(): In Left.
			return rowColToOffsetRec(c.Left, row, col)
		} else {
			// row > leftLines. In Right.
			// But we need to account for the "shared" line?
			// If Left ends with \n, then Line `leftLines` starts clean in Right.
			// If Left does NOT end with \n, then Line `leftLines` is shared.
			// This requires knowing if Left ends with newline?
			// core.Node doesn't expose `EndsWithNewline`.

			// FIXME: This suggests `Lines()` metric alone is insufficient for precise O(log N) line seeking without peeking.
			// However, usually we can optimistically search.
			// Re-evaluating:
			// Actually, just `leftLines` is enough IF we define Row N as "the Nth line *start*".
			// If Left has 5 newlines, it completes lines 0,1,2,3,4.
			// Line 5 starts in Left (possibly) and continues.

			// Let's look at `index.go`. We can implement `OffsetToRowCol` easily.
			// `RowColToOffset` might be trickier.

			// Let's stick to `OffsetToRowCol` as the primary API for now?
			// The user Plan included both.

			// Let's adjust algorithm:
			// To find Row R:
			// If R < Left.Lines(): Recurse Left(R, col).
			// If R == Left.Lines():
			//    This is the tricky shared line.
			//    It starts in Left (after last \n) and continues in Right.
			//    We need to check if Left satisfies the column?
			//    We need to know length of that last segment in Left.
			//    Let's rely on `OffsetToRowCol` being the main feature or implement this slowly?
			//    Ideally we want both.

			// Let's assume we implement `RowColToOffset` by:
			// 1. Find the offset of the start of Row R.
			// 2. Add `col`.
			// 3. Verify that we didn't cross a newline?

			// How to find "Start of Row R"?
			// function StartOfLine(n, lineIndex) -> offset.
			// If n.Lines() > lineIndex:
			//    Recurse.

			// Let's implement `OffsetToRowCol` first correctly in `index.go`.
			// And put `RowColToOffset` there too.
		}

		// Placeholder for now
		return -1
	}
	return -1
}

// OffsetToRowCol converts a byte offset to a 0-based row and column index.
func OffsetToRowCol(n Node, offset int) (row, col int) {
	if n == nil || offset < 0 {
		return -1, -1
	}
	if offset > n.Len() {
		return -1, -1
	}
	// Simple Recursive Strategy?
	// or iterative stack?
	// Recursive is fine for depth ~50.
	return offsetToRowColRec(n, offset)
}

func offsetToRowColRec(n Node, offset int) (int, int) {
	if l, ok := n.(*core.Leaf); ok {
		s := l.String()
		// Count newlines up to offset
		sub := s[:offset]
		row := strings.Count(sub, "\n")
		// Col is distance from last newline
		lastNewline := strings.LastIndex(sub, "\n")
		col := offset - (lastNewline + 1)
		return row, col
	}

	if c, ok := n.(*core.Concat); ok {
		if offset < c.Left.Len() {
			return offsetToRowColRec(c.Left, offset)
		} else {
			// Offset is in Right
			r, col := offsetToRowColRec(c.Right, offset-c.Left.Len())

			// Adjustment:
			// The row in Right is relative to Right's start.
			// We need to add Left's total lines.
			// BUT, what about the shared line?
			// If Left ends with "ABC", and Right is "DEF", Left.Lines=0. Right.Lines=0.
			// Total "ABCDEF". Offset in Right (D) is 0 (relative).
			// offsetToRowCol(Right, 0) -> (0, 0).
			// Result should be (0, 3).
			// WE MISS THE PARTIAL COLUMN.

			// New requirement:
			// `Node` needs `Lines()` AND `LastLineLen()` ? Or `EndsWithNewline`?
			// Without `LastLineLen`, we assume 0, which is wrong.
			// Doing `ByteAt` backwards is slow.

			// Alternative: `offsetToRowCol` isn't purely additive for columns on the split line.
			// The split line inherits "column offset" from the previous chunk.

			// This is complex.
			// Maybe for Phase 1 of line indexing, we rely on scanning just the split boundary?
			// Or we just traverse?
			// Actually, just traversing is O(N) in worst case (one giant line).
			// If many lines, O(Lines)?

			// Let's implement an O(N) scan first and optimize later?
			// No, "O(log N)" was the request in PLAN.
			// To achieve O(log N), `Concat` MUST cache more info.
			// We need `Nodes` to track: `Lines` (count of \n) and probably `LastLineLen`?
			// If we add `LastLineLen` to `Node`, we can solve this.

			// Let's UPDATE THE PLAN?
			// Or Hack it:
			// If we are on line 0 relative to Right, we add `Left.LastLineLen`.
			// Calculating `Left.LastLineLen` might be O(log N) if we descend rightmost?
			// Yes! We can find the length of the last line of `Left` by descending its right children until we find a newline or hit leaf.
			// That is O(depth) = O(log N).
			// So acceptable cost.

			totalRow := c.Left.Lines() + r
			totalCol := col

			if r == 0 {
				// We are on the "continuation" line.
				// Add length of the last line of Left.
				// If Left ended with \n, its last line len is 0.
				lastLen := lastLineLen(c.Left)
				totalCol += lastLen
			}

			return totalRow, totalCol
		}
	}
	return 0, 0
}

func lastLineLen(n Node) int {
	if l, ok := n.(*core.Leaf); ok {
		s := l.String()
		lastIdx := strings.LastIndex(s, "\n")
		if lastIdx == -1 {
			return len(s)
		}
		return len(s) - (lastIdx + 1)
	}
	if c, ok := n.(*core.Concat); ok {
		// If right child has > 0 lines, its last line is the total last line.
		// If right child has 0 lines, it extends the left child's last line.

		rLines := c.Right.Lines()
		if rLines > 0 {
			return lastLineLen(c.Right)
		}
		// Right is just an extension of Left's last line
		return lastLineLen(c.Left) + c.Right.Len()
	}
	return 0
}
