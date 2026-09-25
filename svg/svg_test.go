package svg

import (
	"math"
	"strings"
	"testing"
)

func mustParse(t *testing.T, doc string) *Drawing {
	t.Helper()
	d, err := Parse([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestPathGrammarCompactNumbers(t *testing.T) {
	// Bootstrap Icons style: run-together numbers, arc flags glued to the next
	// value, relative commands.
	d := mustParse(t, `<svg viewBox="0 0 16 16"><path d="M13.354.646a1.207 1.207 0 0 0-1.708 0L8.5 3.793l-.646-.647a.5.5 0 1 0-.708.708L8.293 5z"/></svg>`)
	sp := d.Elements[0].Subpaths[0]
	first := sp[0]
	if first != (Point{13.354, 0.646}) {
		t.Fatalf("first point %v, want (13.354, 0.646)", first)
	}
	last := sp[len(sp)-1]
	if math.Abs(last.X-8.293) > 1e-9 || math.Abs(last.Y-5) > 1e-9 {
		t.Fatalf("last point %v, want (8.293, 5)", last)
	}
}

func TestViewBoxAndInheritedFillRule(t *testing.T) {
	d := mustParse(t, `<svg viewBox="2 3 10 20" fill-rule="evenodd"><g><rect x="2" y="3" width="1" height="1"/></g><rect style="fill-rule:nonzero" x="4" y="4" width="1" height="1"/></svg>`)
	if d.MinX != 2 || d.MinY != 3 || d.Width != 10 || d.Height != 20 {
		t.Fatalf("box %v %v %v %v, want 2 3 10 20", d.MinX, d.MinY, d.Width, d.Height)
	}
	if !d.Elements[0].EvenOdd || d.Elements[1].EvenOdd {
		t.Fatalf("fill rules %v %v, want inherited evenodd then nonzero from style", d.Elements[0].EvenOdd, d.Elements[1].EvenOdd)
	}
}

func TestHiddenAndEmpty(t *testing.T) {
	d := mustParse(t, `<svg viewBox="0 0 4 4"><g style="display:none"><rect width="4" height="4"/></g><circle cx="2" cy="2" r="1"/></svg>`)
	if len(d.Elements) != 1 || len(d.Elements[0].Subpaths[0]) != ellipseSegments {
		t.Fatalf("want only the circle, got %d elements", len(d.Elements))
	}
	_, err := Parse([]byte(`<svg viewBox="0 0 4 4"><rect width="0" height="4"/></svg>`))
	if err == nil || !strings.Contains(err.Error(), "no drawable") {
		t.Fatalf("err %v, want no drawable shapes", err)
	}
}

// Two squares in one path: the inner one wound the other way is a hole under
// nonzero; wound the same way it is not. Under evenodd it is a hole either way.
func TestFillRules(t *testing.T) {
	outer := "M0 0H8V8H0Z"
	sameWay := "M2 2H6V6H2Z"
	otherWay := "M2 2V6H6V2Z"
	tests := []struct {
		name  string
		d     string
		rule  string
		inner uint8
	}{
		{"nonzero opposite winding", outer + otherWay, "", 0},
		{"nonzero same winding", outer + sameWay, "", 255},
		{"evenodd same winding", outer + sameWay, ` fill-rule="evenodd"`, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mustParse(t, `<svg viewBox="0 0 8 8"><path`+tt.rule+` d="`+tt.d+`"/></svg>`).Mask(8)
			got := m.AlphaAt(4, 4).A
			if got != tt.inner {
				t.Errorf("center %d, want %d", got, tt.inner)
			}
			got = m.AlphaAt(0, 0).A
			if got != 255 {
				t.Errorf("ring %d, want 255", got)
			}
		})
	}
}

func TestMaskCoverage(t *testing.T) {
	// Half the viewBox, drawn twice: overlap must not add up past full.
	m := mustParse(t, `<svg viewBox="0 0 4 4"><rect width="2" height="4"/><rect width="2" height="4"/></svg>`).Mask(4)
	for y := range 4 {
		for x := range 4 {
			want := uint8(0)
			if x < 2 {
				want = 255
			}
			got := m.AlphaAt(x, y).A
			if got != want {
				t.Fatalf("(%d,%d) = %d, want %d", x, y, got, want)
			}
		}
	}
	// A rect ending mid-pixel gives partial coverage, not all or nothing.
	m = mustParse(t, `<svg viewBox="0 0 4 4"><rect width="1.5" height="4"/></svg>`).Mask(4)
	got := m.AlphaAt(1, 0).A
	if got != 127 {
		t.Fatalf("half-covered pixel %d, want 127", got)
	}
}

func TestMaskKeepsAspectCentered(t *testing.T) {
	m := mustParse(t, `<svg viewBox="0 0 2 4"><rect width="2" height="4"/></svg>`).Mask(4)
	for x, want := range []uint8{0, 255, 255, 0} {
		got := m.AlphaAt(x, 2).A
		if got != want {
			t.Fatalf("x=%d: %d, want %d (tall drawing centered)", x, got, want)
		}
	}
}
