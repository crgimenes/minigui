package minigui

import (
	"image/color"
	"math"

	"github.com/crgimenes/minigui/svg"
	"github.com/hajimehoshi/ebiten/v2"
)

// Icon is a vector drawing shown at the size of the current text and tinted
// with its color, so one SVG serves every scale and both light and dark
// styles. Bootstrap Icons (fill="currentColor", viewBox 0 0 16 16) work as
// they come.
type Icon struct {
	d   *svg.Drawing
	img *ebiten.Image
	px  int
}

func NewIcon(data []byte) (*Icon, error) {
	d, err := svg.Parse(data)
	if err != nil {
		return nil, err
	}
	return &Icon{d: d}, nil
}

// MustIcon is NewIcon for icons embedded in the program, where a bad file is
// a build mistake rather than a condition to handle.
func MustIcon(data []byte) *Icon {
	ic, err := NewIcon(data)
	if err != nil {
		panic(err)
	}
	return ic
}

// image rasterizes on first use and again only when the size changes. It
// runs from Render, so laying out widgets never touches the GPU.
func (ic *Icon) image(px int) *ebiten.Image {
	if ic.img != nil && ic.px == px {
		return ic.img
	}
	if ic.img != nil {
		ic.img.Deallocate()
	}
	mask := ic.d.Mask(px)
	// White at the mask's coverage, premultiplied; ColorScale tints it.
	pix := make([]byte, 0, px*px*4)
	for _, a := range mask.Pix {
		pix = append(pix, a, a, a, a)
	}
	ic.img = ebiten.NewImage(px, px)
	ic.img.WritePixels(pix)
	ic.px = px
	return ic.img
}

// iconPx is the side of an icon in the current style: the text height.
func (c *Context) iconPx() int {
	return int(math.Round(c.fontH()))
}

// IconButton is Button with an icon before the label. An empty label leaves
// only the icon; give such buttons a meaning that the icon alone carries.
func (c *Context) IconButton(id ID, icon *Icon, label string) bool {
	return c.buttonWith(icon, label, c.style.Button, c.style.Border)
}

// IconToggle is Toggle with an icon before the label.
func (c *Context) IconToggle(id ID, icon *Icon, label string, on bool) bool {
	if on {
		return c.buttonWith(icon, label, c.style.ButtonOn, c.style.Focus)
	}
	return c.buttonWith(icon, label, c.style.Button, c.style.Border)
}

func (c *Context) icon(x, y float64, ic *Icon, px int, col color.RGBA) {
	c.cmds = append(c.cmds, drawCmd{kind: cmdIcon, x: x, y: y, w: float64(px), col: col, icon: ic})
}

func (c *Context) drawIcon(dst *ebiten.Image, cmd *drawCmd) {
	px := int(cmd.w)
	if px < 1 {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(math.Round(cmd.x), math.Round(cmd.y))
	op.ColorScale.ScaleWithColor(cmd.col)
	dst.DrawImage(cmd.icon.image(px), op)
}
