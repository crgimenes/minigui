package minigui

import "testing"

// TestSystemFaceMeasures loads a real system font when one is available and
// checks that a Context measures Latin and CJK text with it. It skips when no
// system font is found (e.g. a headless CI box) rather than failing.
func TestSystemFaceMeasures(t *testing.T) {
	face, err := SystemFace(16)
	if err != nil {
		t.Skipf("no system font available: %v", err)
	}

	var c Context
	c.SetFace(face)

	if w := c.textWidth("AB"); w <= 0 {
		t.Fatalf(`textWidth("AB") = %v, want > 0`, w)
	}
	if w := c.textWidth("日本語"); w <= 0 {
		t.Fatalf("textWidth(CJK) = %v, want > 0", w)
	}
	if rh := c.rowHeight(); rh < rowH {
		t.Fatalf("rowHeight() = %v, want >= %d", rh, rowH)
	}
}
