package minigui

// Touch support maps a finger onto the same pointer fields the widgets already
// read, so a tablet drives buttons, sliders and lists without any widget
// knowing that touch exists.
//
// Ebitengine reports touches separately and never synthesizes mouse input from
// them, so an app that only reads the mouse renders correctly on a tablet and
// then ignores every tap.

// touchState carries the little history that folding a touch onto a mouse
// needs, kept in one place so the merge itself stays testable.
type touchState struct {
	lastX, lastY float64
	seen         bool // a touch has happened, so the mouse position is not to be trusted

	mouseX, mouseY float64
	haveMouse      bool
}

// pointer is the process-wide touch history behind InputFromEbiten. Ebitengine
// calls Update from a single goroutine, and this is only ever reached from
// there, so it needs no lock.
var pointer touchState

// apply folds the current touch into the pointer fields of in.
//
// active says a finger is down this frame, at (x, y); pressed says it went down
// this frame, which is what makes a tap read as a click.
//
// Three details earn their keep. A finger that lifts stops reporting a position
// at all, so the spot it left is held rather than snapping to the origin. Once
// any touch has been seen, the mouse reading is ignored: a device with no mouse
// reports (0, 0) forever, which would otherwise leave whatever widget sits in
// the corner permanently lit. And a mouse that genuinely moves takes the
// pointer back, so a laptop with both a touchscreen and a trackpad keeps
// working with either after the screen has been tapped.
func (t *touchState) apply(in Input, x, y float64, active, pressed bool) Input {
	if t.haveMouse && (in.MouseX != t.mouseX || in.MouseY != t.mouseY) {
		t.seen = false
	}
	t.mouseX, t.mouseY = in.MouseX, in.MouseY
	t.haveMouse = true

	if pressed {
		in.MouseClicked = true
	}
	if active {
		t.lastX, t.lastY = x, y
		t.seen = true
		in.MouseX, in.MouseY = x, y
		in.MouseDown = true
		return in
	}
	if t.seen {
		in.MouseX, in.MouseY = t.lastX, t.lastY
	}
	return in
}
