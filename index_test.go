package rope

import (
	"testing"
)

func TestLineIndexing(t *testing.T) {
	t.Run("Basic Concatenation", func(t *testing.T) {
		// Setup:
		// "Hello\n" (6)
		// "World"   (5) -> Joined: "Hello\nWorld" (11 chars, 1 newline)
		// "!\n"     (2) -> Joined: "Hello\nWorld!\n" (13 chars, 2 newlines)
		// "Test"    (4) -> Joined: "Hello\nWorld!\nTest" (17 chars, 2 newlines)

		r1 := New("Hello\n")
		r2 := New("World")
		r3 := New("!\n")
		r4 := New("Test")

		rope := Join(r1, r2)
		rope = Join(rope, r3)
		rope = Join(rope, r4)

		str := rope.String()
		expectedStr := "Hello\nWorld!\nTest"
		if str != expectedStr {
			t.Fatalf("String mismatch: %q vs %q", str, expectedStr)
		}

		if rope.Lines() != 2 {
			t.Errorf("Lines() got %d, want 2", rope.Lines())
		}

		// Test Conversions
		tests := []struct {
			offset int
			row    int
			col    int
		}{
			{0, 0, 0},     // 'H'
			{5, 0, 5},     // '\n'
			{6, 1, 0},     // 'W'
			{11, 1, 5},    // '!'
			{12, 1, 6},    // '\n'
			{13, 2, 0},    // 'T'
			{16, 2, 3},    // 't'
			{17, 2, 4},    // End
			{-1, -1, -1},  // Invalid
			{100, -1, -1}, // Out of bounds
		}

		for _, tc := range tests {
			// Offset -> RowCol
			r, c := OffsetToRowCol(rope, tc.offset)
			if r != tc.row || c != tc.col {
				t.Errorf("OffsetToRowCol(%d): got (%d, %d), want (%d, %d)", tc.offset, r, c, tc.row, tc.col)
			}

			// RowCol -> Offset (only valid ones)
			if tc.offset >= 0 && tc.offset <= rope.Len() {
				off := RowColToOffset(rope, tc.row, tc.col)
				if off != tc.offset {
					// Special case: RowColToOffset strictness.
					t.Errorf("RowColToOffset(%d, %d): got %d, want %d", tc.row, tc.col, off, tc.offset)
				}
			}
		}
	})

	t.Run("Split Line Logic", func(t *testing.T) {
		// Case: Line spans multiple nodes "A" + "B\n"
		r1 := New("A")
		r2 := New("B\n")
		rope := Join(r1, r2) // "AB\n"

		// Lines should be 1
		if rope.Lines() != 1 {
			t.Errorf("Expected 1 line, got %d", rope.Lines())
		}

		// Offset 1 ('B') -> Row 0, Col 1
		r, c := OffsetToRowCol(rope, 1)
		if r != 0 || c != 1 {
			t.Errorf("OffsetToRowCol(1): got (%d, %d), want (0, 1)", r, c)
		}

		// Row 0, Col 1 -> Offset 1
		off := RowColToOffset(rope, 0, 1)
		if off != 1 {
			t.Errorf("RowColToOffset(0, 1): got %d, want 1", off)
		}
	})

	t.Run("Deeply Nested with Splits", func(t *testing.T) {
		// "Line0\n" + "Line1Start..." + "...Line1End\n" + "Line2"
		r1 := New("Line0\n")
		r2 := New("Line1Start...")
		r3 := New("...Line1End\n")
		r4 := New("Line2")

		// (r1 + (r2 + r3)) + r4
		inner := Join(r2, r3) // "Line1Start......Line1End\n" (Lines=1)
		rope := Join(Join(r1, inner), r4)

		if rope.Lines() != 2 {
			t.Errorf("Expected 2 lines (Line0, Line1), got %d", rope.Lines())
		}

		// Test index inside the split line (Line 1)
		// Offset of start of Line 1 is len("Line0\n") = 6
		// We want to target the first char of "...Line1End\n" which is in r3
		// r2 len is 13 ("Line1Start...").
		// Target char is at global offset 6 + 13 = 19.
		// Row should be 1. Col should be 13.
		r, c := OffsetToRowCol(rope, 19)
		if r != 1 || c != 13 {
			t.Errorf("OffsetToRowCol(19): got (%d, %d), want (1, 13)", r, c)
		}

		off := RowColToOffset(rope, 1, 13)
		if off != 19 {
			t.Errorf("RowColToOffset(1, 13): got %d, want 19", off)
		}
	})
}

func TestLineIndexing_EdgeCases(t *testing.T) {
	// Empty
	empty := New("")
	if empty.Lines() != 0 {
		t.Errorf("Empty lines got %d", empty.Lines())
	}
	r, c := OffsetToRowCol(empty, 0)
	if r != 0 || c != 0 {
		t.Errorf("Empty OffsetToRowCol(0): got (%d,%d)", r, c)
	}

	// Just newline
	nl := New("\n")
	if nl.Lines() != 1 {
		t.Errorf("Newline lines got %d", nl.Lines())
	}

	// Multiple
	multi := New("\n\n\n")
	if multi.Lines() != 3 {
		t.Errorf("Multi lines got %d", multi.Lines())
	}
	// Check middle
	r, c = OffsetToRowCol(multi, 1) // After first \n
	if r != 1 || c != 0 {
		t.Errorf("Multi(1) got (%d,%d)", r, c)
	}
}
