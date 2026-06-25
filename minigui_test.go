package minigui

import "testing"

func TestEditTextInsertAndBackspace(t *testing.T) {
	s, c, a := editText("", 0, 0, Input{Chars: []rune("ab")})
	if s != "ab" || c != 2 {
		t.Fatalf("insert: got %q caret %d", s, c)
	}
	s, c, _ = editText(s, c, a, Input{Backspace: true})
	if s != "a" || c != 1 {
		t.Fatalf("backspace: got %q caret %d", s, c)
	}
}

func TestEditTextInsertsAtCaret(t *testing.T) {
	s, c, _ := editText("ac", 1, 1, Input{Chars: []rune("b")})
	if s != "abc" || c != 2 {
		t.Fatalf("mid insert: got %q caret %d", s, c)
	}
}

func TestEditTextCaretMovement(t *testing.T) {
	if _, c, _ := editText("hello", 5, 5, Input{Left: true}); c != 4 {
		t.Fatalf("left caret %d", c)
	}
	if _, c, _ := editText("hello", 0, 0, Input{Right: true}); c != 1 {
		t.Fatalf("right caret %d", c)
	}
	if _, c, _ := editText("hello", 2, 2, Input{Home: true}); c != 0 {
		t.Fatalf("home caret %d", c)
	}
	if _, c, _ := editText("hello", 2, 2, Input{End: true}); c != 5 {
		t.Fatalf("end caret %d", c)
	}
}

func TestEditTextSkipsControlChars(t *testing.T) {
	if s, _, _ := editText("", 0, 0, Input{Chars: []rune{'\n', '\t', 'x'}}); s != "x" {
		t.Fatalf("control chars not skipped: %q", s)
	}
}

func TestButtonClick(t *testing.T) {
	var c Context

	c.Begin(Input{MouseX: 5, MouseY: 5, MouseClicked: true}, 0, 0)
	clicked := c.Button("b", "go")
	c.End()
	if !clicked {
		t.Fatal("button under the cursor should report a click")
	}

	c.Begin(Input{MouseX: 500, MouseY: 500, MouseClicked: true}, 0, 0)
	if c.Button("b", "go") {
		t.Fatal("button should not click when the cursor is away")
	}
	c.End()
}

func TestTextFieldFocusEditAndDefocus(t *testing.T) {
	var c Context
	s := ""

	// Click on the field to focus it.
	c.Begin(Input{MouseX: 5, MouseY: 5, MouseClicked: true}, 0, 0)
	c.TextField("f", &s)
	c.End()
	if !c.HasFocus() {
		t.Fatal("field should be focused after a click")
	}

	// Type without clicking: focus persists and the text changes.
	c.Begin(Input{Chars: []rune("hi")}, 0, 0)
	changed := c.TextField("f", &s)
	c.End()
	if !changed || s != "hi" {
		t.Fatalf("typed text not applied: changed=%v s=%q", changed, s)
	}

	// Click outside any field clears focus.
	c.Begin(Input{MouseX: 500, MouseY: 500, MouseClicked: true}, 0, 0)
	c.TextField("f", &s)
	c.End()
	if c.HasFocus() {
		t.Fatal("clicking outside a field should clear focus")
	}
}

func TestSameLinePlacesButtonsSideBySide(t *testing.T) {
	var c Context
	// Button "AA" spans x[0,24); SameLine puts "BB" at x[28,52). A click at x=30
	// must hit B, not A.
	c.Begin(Input{MouseX: 30, MouseY: 5, MouseClicked: true}, 0, 0)
	aClicked := c.Button("a", "AA")
	c.SameLine()
	bClicked := c.Button("b", "BB")
	c.End()

	if aClicked {
		t.Fatal("button A should not be clicked at x=30")
	}
	if !bClicked {
		t.Fatal("button B placed by SameLine should be clicked at x=30")
	}
}

func TestToggleReportsClick(t *testing.T) {
	var c Context
	c.Begin(Input{MouseX: 5, MouseY: 5, MouseClicked: true}, 0, 0)
	clicked := c.Toggle("t", "On", true)
	c.End()
	if !clicked {
		t.Fatal("toggle under the cursor should report a click")
	}
}

func TestSetItemWidthFixesButtonWidth(t *testing.T) {
	var c Context
	// "Hi" auto-sizes to 24px; a fixed width of 80 makes it clickable out to x=60.
	c.Begin(Input{MouseX: 60, MouseY: 5, MouseClicked: true}, 0, 0)
	c.SetItemWidth(80)
	clicked := c.Button("b", "Hi")
	c.End()
	if !clicked {
		t.Fatal("button with fixed width 80 should be clickable at x=60")
	}
}
