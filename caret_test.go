package minigui

import "testing"

// Editing counts runes while the text package indexes bytes, so this conversion
// is where a multi-byte character would get split.
func TestByteIndexOfRune(t *testing.T) {
	cases := []struct {
		name  string
		s     string
		runes int
		want  int
	}{
		{"start", "abc", 0, 0},
		{"ascii middle", "abc", 2, 2},
		{"ascii end", "abc", 3, 3},
		{"past the end clamps", "abc", 9, 3},
		{"negative clamps", "abc", -1, 0},
		{"after a two-byte rune", "ação", 1, 1},
		// "ação" is a(1) ç(2) ã(2) o(1), so the fourth rune starts at byte 5.
		{"after two two-byte runes", "ação", 3, 5},
		{"end of accented word", "ação", 4, 6},
		{"after a four-byte rune", "🛩ab", 1, 4},
		{"end with emoji", "🛩ab", 3, 6},
		{"empty", "", 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := byteIndexOfRune(tc.s, tc.runes)
			if got != tc.want {
				t.Errorf("byteIndexOfRune(%q, %d) = %d, want %d", tc.s, tc.runes, got, tc.want)
			}
		})
	}
}

// Without a face the toolkit falls back to a fixed cell, and the caret has to
// land on cell boundaries counted in runes rather than bytes.
func TestCaretXFixedCell(t *testing.T) {
	var c Context

	got := c.caretX("ação", 4)
	want := 4 * float64(charW)
	if got != want {
		t.Errorf("caretX = %v, want %v: four runes should be four cells, not six bytes", got, want)
	}
	if c.caretX("abc", 0) != 0 {
		t.Errorf("caretX at the start = %v, want 0", c.caretX("abc", 0))
	}
}

// The real shaping path only runs with a face loaded, so this skips where no
// system font exists rather than failing.
func TestCaretXWithSystemFace(t *testing.T) {
	face, err := SystemFace(16)
	if err != nil {
		t.Skipf("no system font available: %v", err)
	}

	var c Context
	c.SetFace(face)

	const s = "Wave ação 日本"
	runes := []rune(s)

	if got := c.caretX(s, 0); got != 0 {
		t.Errorf("caretX at the start = %v, want 0", got)
	}

	// The text package documents the caret at the end of a line as the same
	// value as the full line width, which is what every layout here assumes.
	full := c.textWidth(s)
	if got := c.caretX(s, len(runes)); got != full {
		t.Errorf("caretX at the end = %v, want the line width %v", got, full)
	}

	prev := 0.0
	for i := 1; i <= len(runes); i++ {
		cur := c.caretX(s, i)
		if cur < prev {
			t.Fatalf("caretX went backwards at rune %d: %v then %v", i, prev, cur)
		}
		prev = cur
	}

	// An index past the end is clamped instead of panicking, which AdvanceAt
	// would do on an out-of-range byte offset.
	if got := c.caretX(s, len(runes)+50); got != full {
		t.Errorf("caretX past the end = %v, want it clamped to %v", got, full)
	}
}
