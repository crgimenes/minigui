package minigui

import "testing"

func TestFocusSetsAndSurvivesEnd(t *testing.T) {
	var c Context
	s := ""

	// A click far from any field (e.g. on a button that opened a window) would
	// normally clear focus in End; Focus must keep the field focused.
	c.Begin(Input{MouseX: 999, MouseY: 999, MouseClicked: true, MouseDown: true}, 0, 0)
	c.Focus("cmd")
	c.TextField("cmd", &s)
	c.End()

	if !c.HasFocus() || c.focus != "cmd" {
		t.Fatalf("focus = %q, want \"cmd\" still focused after End", c.focus)
	}

	// And typing on the next frame goes into the focused field.
	c.Begin(Input{Chars: []rune("hi")}, 0, 0)
	c.TextField("cmd", &s)
	c.End()
	if s != "hi" {
		t.Fatalf("typed into focused field = %q, want \"hi\"", s)
	}
}
