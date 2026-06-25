package minigui

import "testing"

func TestLineStartEnd(t *testing.T) {
	r := []rune("ab\ncd")
	if got := lineStart(r, 4); got != 3 {
		t.Errorf("lineStart(4) = %d, want 3", got)
	}
	if got := lineStart(r, 1); got != 0 {
		t.Errorf("lineStart(1) = %d, want 0", got)
	}
	if got := lineEnd(r, 0); got != 2 {
		t.Errorf("lineEnd(0) = %d, want 2", got)
	}
	if got := lineEnd(r, 3); got != 5 {
		t.Errorf("lineEnd(3) = %d, want 5", got)
	}
}

func TestCaretUpDownKeepsColumn(t *testing.T) {
	r := []rune("ab\ncd")
	if got := caretDown(r, 1); got != 4 { // line0 col1 -> line1 col1
		t.Errorf("caretDown(1) = %d, want 4", got)
	}
	if got := caretUp(r, 4); got != 1 { // and back
		t.Errorf("caretUp(4) = %d, want 1", got)
	}
	if got := caretUp(r, 1); got != 1 { // first line: no move
		t.Errorf("caretUp(1) = %d, want 1", got)
	}
}

func TestCaretDownClampsColumn(t *testing.T) {
	// "hello\nhi": going down from col 4 lands at the end of the short next line.
	r := []rune("hello\nhi")
	if got := caretDown(r, 4); got != 8 { // line1 has len 2 -> end at index 8
		t.Errorf("caretDown(4) = %d, want 8", got)
	}
}

func TestEditTextAreaEnterInsertsNewline(t *testing.T) {
	s, caret, _ := editTextArea("ab", 2, 2, Input{Enter: true})
	if s != "ab\n" || caret != 3 {
		t.Fatalf("enter: got %q caret %d, want \"ab\\n\" 3", s, caret)
	}
}

func TestEditTextAreaHomeEndLineRelative(t *testing.T) {
	if _, caret, _ := editTextArea("ab\ncd", 4, 4, Input{Home: true}); caret != 3 {
		t.Errorf("home on line 2: caret %d, want 3", caret)
	}
	if _, caret, _ := editTextArea("ab\ncd", 3, 3, Input{End: true}); caret != 5 {
		t.Errorf("end on line 2: caret %d, want 5", caret)
	}
}

func TestEditTextAreaUpExtendsWithShift(t *testing.T) {
	_, caret, anchor := editTextArea("ab\ncd", 4, 4, Input{Up: true, Shift: true})
	if caret != 1 || anchor != 4 {
		t.Fatalf("shift+up: caret=%d anchor=%d, want 1/4", caret, anchor)
	}
}

func TestMultilineRunesKeepsNewlines(t *testing.T) {
	if got := string(multilineRunes("a\nb\tc")); got != "a\nbc" {
		t.Fatalf("multilineRunes = %q, want \"a\\nbc\"", got)
	}
}
