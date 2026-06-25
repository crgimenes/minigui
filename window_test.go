package minigui

import "testing"

func buildWindow(c *Context, w *Window, in Input) {
	c.Begin(in, 0, 0)
	if c.BeginWindow(w) {
		c.Label("content")
		c.EndWindow()
	}
	c.End()
}

func TestWindowDragFollowsMouse(t *testing.T) {
	var c Context
	w := Window{Title: "W", X: 50, Y: 50, Open: true}

	// Press on the title bar starts the drag (grab offset 10,5).
	buildWindow(&c, &w, Input{MouseX: 60, MouseY: 55, MouseClicked: true, MouseDown: true})
	if !c.Dragging() {
		t.Fatal("press on the title bar should start a drag")
	}

	// Hold and move: the window follows by the same offset.
	buildWindow(&c, &w, Input{MouseX: 80, MouseY: 75, MouseDown: true})
	if w.X != 70 || w.Y != 70 {
		t.Fatalf("window did not follow drag: (%v,%v), want (70,70)", w.X, w.Y)
	}

	// Release ends the drag.
	buildWindow(&c, &w, Input{MouseX: 80, MouseY: 75})
	if c.Dragging() {
		t.Fatal("releasing should end the drag")
	}
}

func TestWindowCloseBox(t *testing.T) {
	var c Context
	w := Window{Title: "W", X: 0, Y: 0, Open: true}

	// First frame just to learn the window rectangle.
	c.Begin(Input{}, 0, 0)
	c.BeginWindow(&w)
	c.Label("content")
	rect := c.EndWindow()
	c.End()

	// Click the close box (top-right titleH square).
	titleH := DefaultStyle().RowH
	cx := float64(rect.Max.X) - titleH/2
	cy := float64(rect.Min.Y) + titleH/2
	buildWindow(&c, &w, Input{MouseX: cx, MouseY: cy, MouseClicked: true, MouseDown: true})
	if w.Open {
		t.Fatal("clicking the close box should close the window")
	}
}

func TestWindowNoCloseStaysOpen(t *testing.T) {
	var c Context
	w := Window{Title: "W", X: 0, Y: 0, Open: true, NoClose: true}

	c.Begin(Input{}, 0, 0)
	c.BeginWindow(&w)
	c.Label("content")
	rect := c.EndWindow()
	c.End()

	// A click where the close box would be must not close a NoClose window.
	titleH := DefaultStyle().RowH
	cx := float64(rect.Max.X) - titleH/2
	cy := float64(rect.Min.Y) + titleH/2
	buildWindow(&c, &w, Input{MouseX: cx, MouseY: cy, MouseClicked: true, MouseDown: true})
	if !w.Open {
		t.Fatal("a NoClose window should stay open")
	}
}

func TestClosedWindowSkipsContents(t *testing.T) {
	var c Context
	w := Window{Title: "W", Open: false}
	c.Begin(Input{}, 0, 0)
	if c.BeginWindow(&w) {
		t.Fatal("BeginWindow should return false for a closed window")
	}
	c.End()
}
