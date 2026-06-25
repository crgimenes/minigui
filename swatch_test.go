package minigui

import (
	"image/color"
	"testing"
)

func TestSwatchReportsClick(t *testing.T) {
	var c Context
	c.Begin(Input{MouseX: 5, MouseY: 5, MouseClicked: true}, 0, 0)
	clicked := c.Swatch("sw", color.RGBA{R: 0xff, A: 0xff}, false)
	c.End()
	if !clicked {
		t.Fatal("swatch under the cursor should report a click")
	}
}
