package minigui

import "testing"

func TestPanelSizesToContent(t *testing.T) {
	var c Context
	c.Begin(Input{}, 0, 0)
	c.BeginPanel("Title", 10, 10)
	c.Button("b", "a button")
	rect := c.EndPanel()
	c.End()

	if rect.Min.X != 10 || rect.Min.Y != 10 {
		t.Fatalf("panel origin = %v, want (10,10)", rect.Min)
	}
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		t.Fatalf("panel has empty size: %v", rect)
	}
	// The panel must enclose at least the title bar plus one widget row.
	if want := int(2 * DefaultStyle().RowH); rect.Dy() < want {
		t.Fatalf("panel height %d, want >= %d (title + row)", rect.Dy(), want)
	}
}
