package minigui

import "testing"

func TestStyleInjectionMetrics(t *testing.T) {
	var c Context
	s := DefaultStyle()
	s.RowH = 40
	c.SetStyle(s)

	c.Begin(Input{}, 0, 0)
	if got := c.rowHeight(); got != 40 {
		t.Fatalf("rowHeight() = %v, want 40 (injected RowH)", got)
	}
}

func TestStyleInjectionFieldWidth(t *testing.T) {
	var c Context
	s := DefaultStyle()
	s.FieldW = 320
	c.SetStyle(s)

	str := ""
	// x=300 is past the default 200 width but inside 320, so the field is hit.
	c.Begin(Input{MouseX: 300, MouseY: 5, MouseClicked: true}, 0, 0)
	c.TextField("f", &str)
	c.End()
	if !c.HasFocus() {
		t.Fatal("a field widened to 320 should be focused by a click at x=300")
	}
}

func TestZeroValueContextUsesDefaultStyle(t *testing.T) {
	var c Context
	c.Begin(Input{}, 0, 0)
	if got, want := c.rowHeight(), DefaultStyle().RowH; got != want {
		t.Fatalf("zero-value rowHeight() = %v, want %v", got, want)
	}
}
