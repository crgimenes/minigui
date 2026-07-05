package minigui

// Slider geometry in logical pixels: the visible track is a thin line centered
// in the row, and the knob is a circle riding on it.
const (
	sliderTrackH = 2
	sliderKnobR  = 7
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

	// Track line, filled up to the value, then the knob circle centered on it.
	t := sliderT(*val, lo, hi)
	trackY := y + (h-sliderTrackH)/2
	c.fill(x, trackY, w, sliderTrackH, c.style.Border)
	c.fill(x, trackY, t*w, sliderTrackH, c.style.ButtonOn)

	knob := c.style.Text
	if hot || dragging {
		knob = c.style.Focus
	}
	c.circle(x+t*w, y+h/2, sliderKnobR, knob)

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
