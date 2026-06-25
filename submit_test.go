package minigui

import "testing"

func TestSubmittedOnEnterWhileFocused(t *testing.T) {
	var c Context
	s := ""

	// Click to focus the field; no Enter yet.
	c.Begin(Input{MouseX: 5, MouseY: 5, MouseClicked: true}, 0, 0)
	c.TextField("cmd", &s)
	got := c.Submitted("cmd")
	c.End()
	if got {
		t.Fatal("Submitted should be false without Enter")
	}

	// Press Enter while still focused.
	c.Begin(Input{Enter: true}, 0, 0)
	c.TextField("cmd", &s)
	got = c.Submitted("cmd")
	c.End()
	if !got {
		t.Fatal("Submitted should be true on Enter while focused")
	}

	// Enter without focus does not submit.
	c.Begin(Input{Enter: true}, 0, 0)
	got = c.Submitted("other")
	c.End()
	if got {
		t.Fatal("Submitted should be false for an unfocused id")
	}
}
