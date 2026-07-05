package minigui

import (
	"fmt"
	"testing"
)

func TestListSelectByClick(t *testing.T) {
	var c Context
	items := []string{"a", "b", "c"}
	sel := 0

	// Click within the second row (y in [rowH, 2*rowH)).
	c.Begin(Input{MouseX: 5, MouseY: rowH + 5, MouseClicked: true}, 0, 0)
	changed := c.List("L", items, &sel)
	c.End()

	if !changed || sel != 1 {
		t.Fatalf("click row 1: changed=%v sel=%d, want true/1", changed, sel)
	}
}

func TestListWheelScrollThenSelect(t *testing.T) {
	var c Context
	items := make([]string, 10) // listRows=6, so maxOff=4
	for i := range items {
		items[i] = fmt.Sprintf("i%d", i)
	}
	sel := 0

	// Frame 1: scroll down by 3 while hovering the list.
	c.Begin(Input{MouseX: 5, MouseY: 5, WheelY: -3}, 0, 0)
	c.List("L", items, &sel)
	c.End()

	// Frame 2: click the top visible row, which is now item index 3.
	c.Begin(Input{MouseX: 5, MouseY: 5, MouseClicked: true}, 0, 0)
	changed := c.List("L", items, &sel)
	c.End()

	if !changed || sel != 3 {
		t.Fatalf("after scroll, click top row: changed=%v sel=%d, want true/3", changed, sel)
	}
}

func TestListScrollClampsToEnd(t *testing.T) {
	var c Context
	items := make([]string, 10) // maxOff=4
	for i := range items {
		items[i] = fmt.Sprintf("i%d", i)
	}
	sel := 0

	c.Begin(Input{MouseX: 5, MouseY: 5, WheelY: -100}, 0, 0) // overscroll
	c.List("L", items, &sel)
	c.End()

	got := c.scroll["L"]
	if got != 4 {
		t.Fatalf("scroll clamped to %v, want 4 (maxOff)", got)
	}
}

func TestListEmptyDoesNotPanic(t *testing.T) {
	var c Context
	sel := -1
	c.Begin(Input{MouseX: 5, MouseY: 5, MouseClicked: true}, 0, 0)
	if c.List("L", nil, &sel) {
		t.Fatal("empty list should not report a selection change")
	}
	c.End()
}

func TestListScrollsExternalSelectionIntoView(t *testing.T) {
	var c Context
	items := make([]string, 20)
	for i := range items {
		items[i] = string(rune('a' + i))
	}
	sel := 0

	// Frame 1 records the selection; the list rests at the top.
	c.Begin(Input{}, 0, 0)
	c.List("l", items, &sel)
	c.End()

	// The caller moves the selection far below the fold; the next frame must
	// scroll it into view (offset shows rows 10..15).
	sel = 15
	c.Begin(Input{}, 0, 0)
	c.List("l", items, &sel)
	c.End()
	if got := c.scroll["l"]; got != 10 {
		t.Fatalf("scroll after external selection: got %v, want 10", got)
	}

	// Moving it back above the window scrolls up to it.
	sel = 2
	c.Begin(Input{}, 0, 0)
	c.List("l", items, &sel)
	c.End()
	if got := c.scroll["l"]; got != 2 {
		t.Fatalf("scroll after selecting above: got %v, want 2", got)
	}

	// Wheel scrolling with a steady selection is not snapped back.
	c.Begin(Input{MouseX: 5, MouseY: 5, WheelY: -6}, 0, 0)
	c.List("l", items, &sel)
	c.End()
	if got := c.scroll["l"]; got != 8 {
		t.Fatalf("wheel scroll should stick: got %v, want 8", got)
	}
}

func TestListFirstFrameShowsPresetSelection(t *testing.T) {
	var c Context
	items := make([]string, 20)
	for i := range items {
		items[i] = string(rune('a' + i))
	}

	// The selection exists before the list's first frame (e.g. a document
	// opened from the command line); the very first frame must show it.
	sel := 15
	c.Begin(Input{}, 0, 0)
	c.List("l", items, &sel)
	c.End()
	if got := c.scroll["l"]; got != 10 {
		t.Fatalf("first-frame scroll: got %v, want 10", got)
	}
}
