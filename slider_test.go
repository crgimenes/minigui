package minigui

import "testing"

// Default geometry: the slider row starts at (0, 0), the track spans FieldW
// (200) and the row is RowH (22) tall, so x=150 is 3/4 of the track.

func TestSliderClickJumpsToValue(t *testing.T) {
	var c Context
	v := 0.0

	c.Begin(Input{MouseX: 150, MouseY: 11, MouseDown: true, MouseClicked: true}, 0, 0)
	changed := c.Slider("s", &v, 0, 100)
	c.End()
	if !changed || v != 75 {
		t.Fatalf("click at 3/4: changed=%v v=%v", changed, v)
	}
}

func TestSliderDragFollowsAndReleases(t *testing.T) {
	var c Context
	v := 0.0

	c.Begin(Input{MouseX: 100, MouseY: 11, MouseDown: true, MouseClicked: true}, 0, 0)
	c.Slider("s", &v, 0, 100)
	c.End()
	if v != 50 {
		t.Fatalf("grab: v=%v", v)
	}

	// The drag keeps following the mouse even off the row, without a click.
	c.Begin(Input{MouseX: 20, MouseY: 300, MouseDown: true}, 0, 0)
	changed := c.Slider("s", &v, 0, 100)
	c.End()
	if !changed || v != 10 {
		t.Fatalf("drag off-row: changed=%v v=%v", changed, v)
	}

	// Releasing the button ends the drag: the value stops following.
	c.Begin(Input{MouseX: 180, MouseY: 11}, 0, 0)
	changed = c.Slider("s", &v, 0, 100)
	c.End()
	if changed || v != 10 {
		t.Fatalf("after release: changed=%v v=%v", changed, v)
	}
}

func TestSliderDragClampsToRange(t *testing.T) {
	var c Context
	v := 50.0

	c.Begin(Input{MouseX: 100, MouseY: 11, MouseDown: true, MouseClicked: true}, 0, 0)
	c.Slider("s", &v, 0, 100)
	c.End()

	c.Begin(Input{MouseX: 500, MouseY: 11, MouseDown: true}, 0, 0)
	c.Slider("s", &v, 0, 100)
	c.End()
	if v != 100 {
		t.Fatalf("beyond right edge: v=%v", v)
	}

	c.Begin(Input{MouseX: -40, MouseY: 11, MouseDown: true}, 0, 0)
	c.Slider("s", &v, 0, 100)
	c.End()
	if v != 0 {
		t.Fatalf("beyond left edge: v=%v", v)
	}
}

func TestSliderIgnoresClickAway(t *testing.T) {
	var c Context
	v := 42.0

	c.Begin(Input{MouseX: 500, MouseY: 500, MouseDown: true, MouseClicked: true}, 0, 0)
	changed := c.Slider("s", &v, 0, 100)
	c.End()
	if changed || v != 42 {
		t.Fatalf("click away: changed=%v v=%v", changed, v)
	}
}

func TestSliderSteadyHoldReportsNoChange(t *testing.T) {
	var c Context
	v := 0.0

	c.Begin(Input{MouseX: 100, MouseY: 11, MouseDown: true, MouseClicked: true}, 0, 0)
	c.Slider("s", &v, 0, 100)
	c.End()

	c.Begin(Input{MouseX: 100, MouseY: 11, MouseDown: true}, 0, 0)
	changed := c.Slider("s", &v, 0, 100)
	c.End()
	if changed || v != 50 {
		t.Fatalf("steady hold: changed=%v v=%v", changed, v)
	}
}

func TestSliderGrabIsExclusive(t *testing.T) {
	var c Context
	a, b := 0.0, 0.0

	// Two sliders stacked; a click on the first must not move the second even
	// though the drag state is checked by both on the same frame.
	c.Begin(Input{MouseX: 100, MouseY: 11, MouseDown: true, MouseClicked: true}, 0, 0)
	c.Slider("a", &a, 0, 100)
	c.Slider("b", &b, 0, 100)
	c.End()
	if a != 50 || b != 0 {
		t.Fatalf("grab exclusivity: a=%v b=%v", a, b)
	}
}

func TestSliderValueMapping(t *testing.T) {
	if got := sliderValue(50, 0, 200, -10, 10); got != -5 {
		t.Fatalf("quarter of a signed range: %v", got)
	}
	if got := sliderValue(999, 0, 0, 3, 9); got != 3 {
		t.Fatalf("degenerate width should pin to lo: %v", got)
	}
	if got := sliderT(5, 5, 5); got != 0 {
		t.Fatalf("degenerate range should pin to 0: %v", got)
	}
	if got := sliderT(200, 0, 100); got != 1 {
		t.Fatalf("out-of-range value should clamp: %v", got)
	}
}
