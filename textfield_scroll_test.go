package minigui

import (
	"strings"
	"testing"
)

func TestTextFieldScrollsToKeepCaretVisible(t *testing.T) {
	var c Context
	s := strings.Repeat("x", 60) // far wider than the field
	c.focus = "f"
	c.caret = 60 // caret at the end

	c.Begin(Input{}, 0, 0)
	c.TextField("f", &s)
	c.End()

	xoff := c.scroll["f"]
	if xoff <= 0 {
		t.Fatalf("a long field should scroll, xoff=%v", xoff)
	}

	inner := float64(fieldW - 2*pad)
	caretView := float64(60)*charW - xoff
	if caretView < 0 || caretView > inner+1e-6 {
		t.Fatalf("caret not kept in view: caretView=%v inner=%v", caretView, inner)
	}
}

func TestTextFieldNoScrollWhenItFits(t *testing.T) {
	var c Context
	s := "short"
	c.focus = "f"
	c.caret = 5

	c.Begin(Input{}, 0, 0)
	c.TextField("f", &s)
	c.End()

	if xoff := c.scroll["f"]; xoff != 0 {
		t.Fatalf("text that fits should not scroll, xoff=%v", xoff)
	}
}
