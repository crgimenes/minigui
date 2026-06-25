package minigui

import "testing"

func TestSelectAllThenType(t *testing.T) {
	// Ctrl/Cmd+A selects everything; typing then replaces the whole selection.
	s, caret, anchor := editText("hello", 2, 2, Input{SelectAll: true})
	if caret != 5 || anchor != 0 {
		t.Fatalf("select all: caret=%d anchor=%d, want 5/0", caret, anchor)
	}
	s, caret, anchor = editText(s, caret, anchor, Input{Chars: []rune("X")})
	if s != "X" || caret != 1 || anchor != 1 {
		t.Fatalf("type over selection: got %q caret=%d anchor=%d", s, caret, anchor)
	}
}

func TestShiftExtendsSelection(t *testing.T) {
	// Shift+Left from the end extends the selection leftward (anchor stays).
	_, caret, anchor := editText("hello", 5, 5, Input{Left: true, Shift: true})
	if caret != 4 || anchor != 5 {
		t.Fatalf("shift+left: caret=%d anchor=%d, want 4/5", caret, anchor)
	}
	// A plain Left then collapses the selection to its low end without moving.
	_, caret, anchor = editText("hello", caret, anchor, Input{Left: true})
	if caret != 4 || anchor != 4 {
		t.Fatalf("collapse left: caret=%d anchor=%d, want 4/4", caret, anchor)
	}
}

func TestBackspaceDeletesSelection(t *testing.T) {
	// Selection [1,4) = "ell"; Backspace removes it and collapses the caret.
	s, caret, anchor := editText("hello", 4, 1, Input{Backspace: true})
	if s != "ho" || caret != 1 || anchor != 1 {
		t.Fatalf("backspace selection: got %q caret=%d anchor=%d, want \"ho\" 1/1", s, caret, anchor)
	}
}

func TestInsertRunesReplacesSelection(t *testing.T) {
	// Replace selection [1,4) of "hello" with "EY" (paste semantics).
	s, caret, anchor := insertRunes("hello", 4, 1, []rune("EY"))
	if s != "hEYo" || caret != 3 || anchor != 3 {
		t.Fatalf("insertRunes: got %q caret=%d anchor=%d, want \"hEYo\" 3/3", s, caret, anchor)
	}
}

func TestInsertRunesNilDeletes(t *testing.T) {
	s, caret, _ := insertRunes("hello", 4, 1, nil)
	if s != "ho" || caret != 1 {
		t.Fatalf("delete via insertRunes: got %q caret=%d, want \"ho\" 1", s, caret)
	}
}

func TestInlineRunesStripsControl(t *testing.T) {
	if got := string(inlineRunes("a\nb\tc")); got != "abc" {
		t.Fatalf("inlineRunes = %q, want \"abc\"", got)
	}
}
