package minigui

// Slider geometry in logical pixels: the visible track is a thin bar centered
// in the row, and the knob is a square riding on it.
const (
	sliderTrackH = 4
	sliderKnob   = 14
)

// Slider draws a horizontal draggable track that edits *val within [lo, hi],
// reporting whether the value changed this frame. Clicking anywhere on the row
// jumps the value to the cursor and starts a drag that keeps following the
// mouse until the button is released, even if the cursor leaves the row. The
// width is the style's FieldW; a label and a formatted value are composed by
// the caller with Label and SameLine, like the other widgets.
func (c *Context) Slider(id ID, val *float64, lo, hi float64) bool {
	w, h := c.style.FieldW, c.rowHeight()
	x, y := c.x, c.y
	hot := within(c.in.MouseX, c.in.MouseY, x, y, w, h)

	if c.dragSlider == "" && hot && c.in.MouseClicked {
		c.dragSlider = id
	}
	dragging := c.dragSlider == id

	changed := false
	if dragging {
		v := sliderValue(c.in.MouseX, x, w, lo, hi)
		if v != *val {
			*val = v
			changed = true
		}
	}

	// Track, filled up to the value, then the knob centered on it.
	t := sliderT(*val, lo, hi)
	trackY := y + (h-sliderTrackH)/2
	c.fill(x, trackY, w, sliderTrackH, c.style.Field)
	c.fill(x, trackY, t*w, sliderTrackH, c.style.ButtonOn)
	c.border(x, trackY, w, sliderTrackH, c.style.Border)

	knobFill := c.style.Button
	if hot || dragging {
		knobFill = c.style.ButtonHot
	}
	knobBorder := c.style.Border
	if dragging {
		knobBorder = c.style.Focus
	}
	kx := x + t*w - sliderKnob/2
	ky := y + (h-sliderKnob)/2
	c.fill(kx, ky, sliderKnob, sliderKnob, knobFill)
	c.border(kx, ky, sliderKnob, sliderKnob, knobBorder)

	c.advance(w, h)
	return changed
}

// sliderValue maps a mouse x on a track starting at x with width w to a value
// in [lo, hi], clamping the cursor to the track's ends.
func sliderValue(mx, x, w, lo, hi float64) float64 {
	if w <= 0 {
		return lo
	}
	t := (mx - x) / w
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return lo + t*(hi-lo)
}

// sliderT maps a value to its normalized position on the track, clamped to
// [0, 1] so an out-of-range value still draws a sane knob.
func sliderT(v, lo, hi float64) float64 {
	if hi == lo {
		return 0
	}
	t := (v - lo) / (hi - lo)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return t
}
