package minigui

import "testing"

func TestCaretAtX(t *testing.T) {
	var c Context // no face: each rune is charW (6px) wide
	cases := []struct {
		x    float64
		want int
	}{
		{-5, 0},
		{0, 0},
		{2, 0},   // within the first cell, nearer the left boundary
		{7, 1},   // within the second cell, nearer its left boundary
		{100, 3}, // past the end clamps to len
	}
	for _, tc := range cases {
		if got := c.caretAtX("abc", tc.x); got != tc.want {
			t.Errorf("caretAtX(%q, %v) = %d, want %d", "abc", tc.x, got, tc.want)
		}
	}
}

func TestWordBounds(t *testing.T) {
	lo, hi := wordBounds("foo bar", 5)
	if lo != 4 || hi != 7 {
		t.Errorf("word at 5: [%d,%d), want [4,7)", lo, hi)
	}
	lo, hi = wordBounds("foo bar", 3) // the space
	if lo != 3 || hi != 4 {
		t.Errorf("space at 3: [%d,%d), want [3,4)", lo, hi)
	}
	if lo, hi := wordBounds("", 0); lo != 0 || hi != 0 {
		t.Errorf("empty: [%d,%d), want [0,0)", lo, hi)
	}
	if lo, hi := wordBounds("hi", 5); lo != 0 || hi != 2 { // pos clamps into range
		t.Errorf("clamped: [%d,%d), want [0,2)", lo, hi)
	}
}
