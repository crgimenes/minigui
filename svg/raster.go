package svg

import (
	"image"
	"math"
	"sort"
)

// superSample is the grid per output pixel. Four is where diagonals stop
// looking stepped at 16px, the smallest size an icon has to survive.
const superSample = 4

// Mask renders the drawing into a size x size coverage mask, the viewBox
// scaled to fit and centered. Elements are unioned; each fills its subpaths
// by its own fill rule.
func (d *Drawing) Mask(size int) *image.Alpha {
	out := image.NewAlpha(image.Rect(0, 0, size, size))
	if size < 1 || d.Width <= 0 || d.Height <= 0 {
		return out
	}
	n := size * superSample
	scale := float64(n) / math.Max(d.Width, d.Height)
	offX := (float64(n) - d.Width*scale) / 2
	offY := (float64(n) - d.Height*scale) / 2
	// One flag per supersample: setting it twice is still one sample, which
	// is what makes overlapping elements a union rather than a sum.
	covered := make([]bool, n*n)
	var xs []crossing
	for _, e := range d.Elements {
		for y := range n {
			yc := float64(y) + 0.5
			xs = xs[:0]
			for _, sp := range e.Subpaths {
				xs = appendCrossings(xs, sp, yc, scale, offX-d.MinX*scale, offY-d.MinY*scale)
			}
			if len(xs) < 2 {
				continue
			}
			sort.Slice(xs, func(i, j int) bool { return xs[i].x < xs[j].x })
			fillSpans(covered[y*n:(y+1)*n], xs, e.EvenOdd)
		}
	}
	for y := range size {
		for x := range size {
			c := 0
			for sy := range superSample {
				row := covered[(y*superSample+sy)*n:]
				for sx := range superSample {
					if row[x*superSample+sx] {
						c++
					}
				}
			}
			out.Pix[y*out.Stride+x] = uint8(c * 255 / (superSample * superSample)) // #nosec G115 -- c <= superSample^2
		}
	}
	return out
}

type crossing struct {
	x   float64
	dir int
}

// appendCrossings adds where scanline yc crosses the closed polygon pts, in
// supersampled pixels. The half-open test keeps a vertex exactly on the line
// from counting twice.
func appendCrossings(xs []crossing, pts []Point, yc, scale, ox, oy float64) []crossing {
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		ay, by := a.Y*scale+oy, b.Y*scale+oy
		if (ay <= yc) == (by <= yc) {
			continue
		}
		ax, bx := a.X*scale+ox, b.X*scale+ox
		dir := 1
		if by < ay {
			dir = -1
		}
		xs = append(xs, crossing{x: ax + (yc-ay)/(by-ay)*(bx-ax), dir: dir})
	}
	return xs
}

// fillSpans marks the samples of one supersampled row inside the element.
func fillSpans(row []bool, xs []crossing, evenOdd bool) {
	wind := 0
	for i := 0; i+1 < len(xs); i++ {
		wind += xs[i].dir
		inside := wind != 0
		if evenOdd {
			inside = i%2 == 0
		}
		if !inside {
			continue
		}
		x0 := max(int(math.Round(xs[i].x)), 0)
		x1 := min(int(math.Round(xs[i+1].x)), len(row))
		for x := x0; x < x1; x++ {
			row[x] = true
		}
	}
}
