package minigui

import "testing"

func TestTouchDrivesPointer(t *testing.T) {
	var s touchState

	got := s.apply(Input{}, 40, 25, true, true)
	if !got.MouseDown || !got.MouseClicked {
		t.Errorf("a finger going down should read as a held, just-clicked button, got down=%v clicked=%v", got.MouseDown, got.MouseClicked)
	}
	if got.MouseX != 40 || got.MouseY != 25 {
		t.Errorf("pointer = (%v, %v), want the touch position (40, 25)", got.MouseX, got.MouseY)
	}
}

func TestTouchHeldReportsNoFurtherClicks(t *testing.T) {
	var s touchState

	s.apply(Input{}, 40, 25, true, true)
	got := s.apply(Input{}, 44, 25, true, false)
	if !got.MouseDown {
		t.Error("a finger still down should keep the button held")
	}
	if got.MouseClicked {
		t.Error("holding a finger should not keep re-reporting a click")
	}
	if got.MouseX != 44 {
		t.Errorf("pointer x = %v, want the finger to have dragged to 44", got.MouseX)
	}
}

// A lifted finger stops reporting a position, so the pointer has to stay where
// it was instead of snapping to the origin.
func TestTouchReleaseHoldsLastPosition(t *testing.T) {
	var s touchState

	s.apply(Input{}, 90, 60, true, true)
	got := s.apply(Input{}, 0, 0, false, false)
	if got.MouseDown {
		t.Error("a lifted finger should release the button")
	}
	if got.MouseX != 90 || got.MouseY != 60 {
		t.Errorf("pointer = (%v, %v), want it held at the release point (90, 60)", got.MouseX, got.MouseY)
	}
}

// A device with no mouse reports (0, 0) forever, which would leave whatever
// widget sits in the corner lit. Once a touch has been seen, that reading is
// ignored.
func TestMouseIgnoredAfterFirstTouch(t *testing.T) {
	var s touchState

	stray := Input{MouseX: 0, MouseY: 0}
	got := s.apply(stray, 0, 0, false, false)
	if got.MouseX != 0 || got.MouseY != 0 {
		t.Error("before any touch the mouse position should pass through untouched")
	}

	s.apply(Input{}, 120, 80, true, true)
	got = s.apply(stray, 0, 0, false, false)
	if got.MouseX != 120 || got.MouseY != 80 {
		t.Errorf("pointer = (%v, %v), want the stray (0, 0) mouse reading ignored in favour of (120, 80)", got.MouseX, got.MouseY)
	}
}

// A laptop with both a touchscreen and a trackpad has to keep working with
// either one, so a mouse that actually moves takes the pointer back from touch.
func TestMovingMouseTakesPointerBackFromTouch(t *testing.T) {
	var s touchState

	s.apply(Input{MouseX: 5, MouseY: 5}, 120, 80, true, true)
	held := s.apply(Input{MouseX: 5, MouseY: 5}, 0, 0, false, false)
	if held.MouseX != 120 || held.MouseY != 80 {
		t.Fatalf("pointer = (%v, %v), want it still held at the touch point while the mouse sits still", held.MouseX, held.MouseY)
	}

	moved := s.apply(Input{MouseX: 61, MouseY: 62}, 0, 0, false, false)
	if moved.MouseX != 61 || moved.MouseY != 62 {
		t.Errorf("pointer = (%v, %v), want the moved mouse (61, 62) to take over", moved.MouseX, moved.MouseY)
	}
}

// Without touch the mouse has to come through exactly as it did before, so the
// desktop behaviour is unchanged.
func TestMousePassesThroughUntouched(t *testing.T) {
	var s touchState

	in := Input{MouseX: 12, MouseY: 34, MouseDown: true, MouseClicked: true}
	got := s.apply(in, 0, 0, false, false)
	if got.MouseX != in.MouseX || got.MouseY != in.MouseY {
		t.Errorf("pointer = (%v, %v), want the mouse position (%v, %v)", got.MouseX, got.MouseY, in.MouseX, in.MouseY)
	}
	if !got.MouseDown || !got.MouseClicked {
		t.Errorf("mouse button state = down %v clicked %v, want both held from the mouse", got.MouseDown, got.MouseClicked)
	}
}
