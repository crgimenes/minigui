package minigui

import "testing"

const squareSVG = `<svg viewBox="0 0 16 16"><rect width="16" height="16"/></svg>`

func TestIconButtonWidthAndClick(t *testing.T) {
	ic := MustIcon([]byte(squareSVG))
	var c Context
	st := c.Style()
	px := float64(c.iconPx())

	c.Begin(Input{MouseX: 2, MouseY: 2, MouseClicked: true}, 0, 0)
	clicked := c.IconButton("b", ic, "go")
	c.End()
	if !clicked {
		t.Fatal("icon button under the cursor should report a click")
	}
	want := 2*st.Pad + px + st.Gap + c.textWidth("go")
	if c.lastW != want {
		t.Fatalf("width %v, want %v (pad, icon, gap, label, pad)", c.lastW, want)
	}

	c.Begin(Input{}, 0, 0)
	c.IconToggle("t", ic, "", true)
	c.End()
	if c.lastW != 2*st.Pad+px {
		t.Fatalf("icon-only width %v, want %v", c.lastW, 2*st.Pad+px)
	}
	n := 0
	for _, cmd := range c.cmds {
		if cmd.kind == cmdText {
			t.Fatal("an icon-only button must draw no label")
		}
		if cmd.kind == cmdIcon {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d icons drawn, want 1", n)
	}
}

func TestNewIconRejectsEmptyDrawing(t *testing.T) {
	_, err := NewIcon([]byte(`<svg viewBox="0 0 16 16"></svg>`))
	if err == nil {
		t.Fatal("an SVG with nothing to draw must not become an icon")
	}
}
