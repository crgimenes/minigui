// Package minigui is a tiny immediate-mode GUI toolkit for Ebitengine. Each
// frame the caller builds an Input, runs Begin, calls widgets, then End; widgets
// process input and append draw commands that Render flushes. Keeping the logic
// separate from rendering makes the toolkit testable without a window and keeps a
// consistent look across the host application.
package minigui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ID identifies a widget across frames, used to track which one holds focus.
type ID string

// Face is the text face used to render widget labels. It is an alias for
// Ebitengine's text/v2 Face, so a caller can supply any font: a system font (see
// SystemFace), an embedded TTF, or a bitmap face. When a Context has no face it
// falls back to Ebitengine's built-in debug font.
type Face = text.Face

// Layout metrics in logical pixels. The v1 toolkit uses Ebitengine's built-in
// debug font, a fixed 6x16 cell, so text and carets are sized from that.
const (
	charW  = 6
	charH  = 16
	rowH   = 22
	pad    = 6
	gap    = 4
	fieldW = 200
)

var (
	colText   = color.RGBA{0xc8, 0xf0, 0xff, 0xff}
	colBorder = color.RGBA{0x40, 0x80, 0x90, 0xff}
	colBtn    = color.RGBA{0x12, 0x20, 0x2a, 0xff}
	colBtnHot = color.RGBA{0x1e, 0x38, 0x44, 0xff}
	colBtnOn  = color.RGBA{0x1e, 0x44, 0x52, 0xff} // active toggle background
	colField  = color.RGBA{0x06, 0x0c, 0x12, 0xff}
	colFocus  = color.RGBA{0x80, 0xff, 0xff, 0xff}
)

// Input is the per-frame input snapshot. It is plain data so the toolkit can be
// driven from tests; InputFromEbiten builds it from the live ebiten state.
type Input struct {
	MouseX, MouseY float64
	MouseDown      bool    // left button held
	MouseClicked   bool    // left button pressed this frame
	WheelY         float64 // vertical scroll this frame (ebiten wheel dy)
	Chars          []rune
	Backspace      bool
	Left, Right    bool
	Home, End      bool
	Enter          bool
}

type cmdKind int

const (
	cmdFill cmdKind = iota
	cmdBorder
	cmdText
)

type drawCmd struct {
	kind       cmdKind
	x, y, w, h float64
	col        color.RGBA
	s          string
	clip       image.Rectangle // text only: clip the glyphs to this rect when set
}

// Context holds immediate-mode state that persists across frames (focus, caret)
// plus this frame's input, layout cursor and draw-command buffer.
type Context struct {
	in           Input
	x0, y0       float64
	x, y         float64
	cmds         []drawCmd
	focus        ID
	caret        int
	clickedField bool           // a field captured the click this frame
	scroll       map[ID]float64 // persisted scroll position per list, in rows

	// Geometry of the last widget, so SameLine can place the next one beside it.
	lastX, lastY, lastW float64

	itemW float64 // fixed width for buttons/toggles (0 = size to text)

	face Face // text face for labels and fields; nil uses the debug font
}

// Begin starts a frame, laying widgets out from the given top-left position.
func (c *Context) Begin(in Input, x, y float64) {
	c.in = in
	c.x0, c.y0 = x, y
	c.x, c.y = x, y
	c.cmds = c.cmds[:0]
	c.clickedField = false
	c.itemW = 0
}

// SetItemWidth fixes the width of subsequent buttons and toggles so a column of
// them can share one size (0 sizes each to its text). Reset every frame by Begin.
func (c *Context) SetItemWidth(w float64) {
	c.itemW = w
}

// SetFace sets the text face used to measure and draw widget labels. Pass nil to
// fall back to Ebitengine's built-in debug font (a fixed 6x16 cell, ASCII only).
// A real face also lets non-Latin text (e.g. Japanese) render. Set it once on the
// Context, e.g. before the first Begin; it persists across frames.
func (c *Context) SetFace(f Face) {
	c.face = f
}

// End finishes the frame, clearing focus when the click missed every field.
func (c *Context) End() {
	if c.in.MouseClicked && !c.clickedField {
		c.focus = ""
	}
}

// HasFocus reports whether a text field owns keyboard input, so the caller can
// suspend single-key shortcuts while the user types.
func (c *Context) HasFocus() bool {
	return c.focus != ""
}

// ClearFocus drops keyboard focus, e.g. when the panel owning the field is hidden.
func (c *Context) ClearFocus() {
	c.focus = ""
}

// SameLine places the next widget to the right of the one just drawn, on the
// same row, instead of on a new line below it.
func (c *Context) SameLine() {
	c.x = c.lastX + c.lastW + gap
	c.y = c.lastY
}

// Label draws a line of static text.
func (c *Context) Label(s string) {
	rh := c.rowHeight()
	c.textAt(c.x, c.y+(rh-c.fontH())/2, s, colText)
	c.advance(c.textWidth(s), rh)
}

// Button draws a clickable button and reports whether it was clicked this frame.
func (c *Context) Button(id ID, label string) bool {
	return c.button(label, colBtn, colBorder)
}

// Toggle is a button that shows an active state, used for tools and on/off
// options; it reports whether it was clicked this frame.
func (c *Context) Toggle(id ID, label string, on bool) bool {
	if on {
		return c.button(label, colBtnOn, colFocus)
	}
	return c.button(label, colBtn, colBorder)
}

// swatchSize is the side of a color swatch in logical pixels.
const swatchSize = 18

// Swatch draws a clickable color square, highlighted when selected, and reports
// whether it was clicked this frame.
func (c *Context) Swatch(id ID, col color.RGBA, selected bool) bool {
	w, h := float64(swatchSize), float64(swatchSize)
	hot := within(c.in.MouseX, c.in.MouseY, c.x, c.y, w, h)

	c.fill(c.x, c.y, w, h, col)
	border := colBorder
	switch {
	case selected:
		border = colFocus
	case hot:
		border = colText
	}
	c.border(c.x, c.y, w, h, border)

	clicked := hot && c.in.MouseClicked
	c.advance(w, h)
	return clicked
}

// button draws a labelled box with the given fill and border, brightening on
// hover, and reports a click. The label is centered; the width sizes to the text
// unless a fixed item width is set (SetItemWidth). Shared core of Button/Toggle.
func (c *Context) button(label string, fill, border color.RGBA) bool {
	textW := c.textWidth(label)
	w := textW + 2*pad
	if c.itemW > w {
		w = c.itemW
	}
	h := c.rowHeight()
	hot := within(c.in.MouseX, c.in.MouseY, c.x, c.y, w, h)
	if hot {
		fill = colBtnHot
	}
	c.fill(c.x, c.y, w, h, fill)
	c.border(c.x, c.y, w, h, border)
	c.textAt(c.x+(w-textW)/2, c.y+(h-c.fontH())/2, label, colText)

	clicked := hot && c.in.MouseClicked
	c.advance(w, h)
	return clicked
}

// Render flushes the frame's draw commands into dst.
func (c *Context) Render(dst *ebiten.Image) {
	for i := range c.cmds {
		cmd := &c.cmds[i]
		switch cmd.kind {
		case cmdFill:
			vector.FillRect(dst, float32(cmd.x), float32(cmd.y), float32(cmd.w), float32(cmd.h), cmd.col, false)
		case cmdBorder:
			vector.StrokeRect(dst, float32(cmd.x), float32(cmd.y), float32(cmd.w), float32(cmd.h), 1, cmd.col, true)
		case cmdText:
			c.drawText(dst, cmd)
		}
	}
}

// drawText renders one text command, clipped to cmd.clip when set. With a face it
// uses text/v2 (honoring the command's color); without one it falls back to
// Ebitengine's debug font, which ignores color.
func (c *Context) drawText(dst *ebiten.Image, cmd *drawCmd) {
	target := dst
	if !cmd.clip.Empty() {
		clip := cmd.clip.Intersect(dst.Bounds())
		if clip.Empty() {
			return
		}
		target = dst.SubImage(clip).(*ebiten.Image)
	}
	if c.face == nil {
		ebitenutil.DebugPrintAt(target, cmd.s, int(cmd.x), int(cmd.y))
		return
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(cmd.x, cmd.y)
	op.ColorScale.ScaleWithColor(cmd.col)
	text.Draw(target, cmd.s, c.face, op)
}

// textWidth returns the rendered width of s in logical pixels under the current
// face, or the fixed-cell width when no face is set.
func (c *Context) textWidth(s string) float64 {
	if c.face == nil {
		return float64(len([]rune(s))) * charW
	}
	return text.Advance(s, c.face)
}

// fontH returns the text height used to vertically center labels: the face's line
// height when set, or the fixed debug-font cell height otherwise.
func (c *Context) fontH() float64 {
	if c.face == nil {
		return charH
	}
	m := c.face.Metrics()
	return m.HAscent + m.HDescent
}

// rowHeight is the height of a standard widget row. Without a face it stays at the
// v1 metric; with one it grows to fit the font's line height.
func (c *Context) rowHeight() float64 {
	if c.face == nil {
		return rowH
	}
	h := c.fontH() + 2*pad
	if h < rowH {
		h = rowH
	}
	return h
}

// advance records the just-drawn widget's geometry (so SameLine can use it) and
// moves the layout cursor to the next row.
func (c *Context) advance(w, h float64) {
	c.lastX = c.x
	c.lastY = c.y
	c.lastW = w
	c.newlineH(h)
}

// newlineH advances the layout cursor to the next row, leaving room for a widget
// of the given height.
func (c *Context) newlineH(h float64) {
	c.y += h + gap
	c.x = c.x0
}

// setScroll stores a list's scroll position, lazily creating the map.
func (c *Context) setScroll(id ID, v float64) {
	if c.scroll == nil {
		c.scroll = map[ID]float64{}
	}
	c.scroll[id] = v
}

func (c *Context) fill(x, y, w, h float64, col color.RGBA) {
	c.cmds = append(c.cmds, drawCmd{kind: cmdFill, x: x, y: y, w: w, h: h, col: col})
}

func (c *Context) border(x, y, w, h float64, col color.RGBA) {
	c.cmds = append(c.cmds, drawCmd{kind: cmdBorder, x: x, y: y, w: w, h: h, col: col})
}

func (c *Context) textAt(x, y float64, s string, col color.RGBA) {
	c.cmds = append(c.cmds, drawCmd{kind: cmdText, x: x, y: y, s: s, col: col})
}

func (c *Context) textClip(x, y float64, s string, col color.RGBA, clip image.Rectangle) {
	c.cmds = append(c.cmds, drawCmd{kind: cmdText, x: x, y: y, s: s, col: col, clip: clip})
}

func within(px, py, x, y, w, h float64) bool {
	return px >= x && px < x+w && py >= y && py < y+h
}

// InputFromEbiten samples the live ebiten input into an Input snapshot.
func InputFromEbiten() Input {
	mx, my := ebiten.CursorPosition()
	in := Input{
		MouseX:       float64(mx),
		MouseY:       float64(my),
		MouseDown:    ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		MouseClicked: inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		Backspace:    repeatKey(ebiten.KeyBackspace),
		Left:         repeatKey(ebiten.KeyArrowLeft),
		Right:        repeatKey(ebiten.KeyArrowRight),
		Home:         inpututil.IsKeyJustPressed(ebiten.KeyHome),
		End:          inpututil.IsKeyJustPressed(ebiten.KeyEnd),
		Enter:        inpututil.IsKeyJustPressed(ebiten.KeyEnter),
	}
	in.Chars = ebiten.AppendInputChars(nil)
	_, wy := ebiten.Wheel()
	in.WheelY = wy
	return in
}

// repeatKey reports a key as pressed on the initial press and then on a steady
// repeat while held, so backspace and the arrow keys feel right while typing.
func repeatKey(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	if d == 1 {
		return true
	}
	return d >= 30 && (d-30)%4 == 0
}
