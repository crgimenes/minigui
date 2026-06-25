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
// Default metrics in logical pixels. charW/charH are the built-in debug font's
// fixed cell, used only when no face is set; the rest seed DefaultStyle and can
// be overridden per Context through a Style.
const (
	charW  = 6
	charH  = 16
	rowH   = 22
	pad    = 6
	gap    = 4
	fieldW = 200
)

// Style holds the look of a Context: the text face, the widget colors and the
// layout metrics. Start from DefaultStyle, override the fields you want, then
// apply it with (*Context).SetStyle, so a host application can give the toolkit
// its own palette and sizing.
type Style struct {
	Face Face // text face; nil uses the built-in debug font

	Text      color.RGBA // label and field text
	Border    color.RGBA // widget outline
	Button    color.RGBA // button/toggle background
	ButtonHot color.RGBA // hovered button background
	ButtonOn  color.RGBA // active toggle background
	Field     color.RGBA // text field and list background
	Focus     color.RGBA // focused outline, caret and selection accent
	Selection color.RGBA // background of selected text

	RowH   float64 // height of a standard widget row
	Pad    float64 // inner padding
	Gap    float64 // gap between adjacent widgets
	FieldW float64 // width of text fields and lists
}

// DefaultStyle returns the toolkit's built-in look: a dark, cyan-tinted scheme
// with the v1 metrics. A zero-value Context uses it until SetStyle is called.
func DefaultStyle() Style {
	return Style{
		Text:      color.RGBA{0xc8, 0xf0, 0xff, 0xff},
		Border:    color.RGBA{0x40, 0x80, 0x90, 0xff},
		Button:    color.RGBA{0x12, 0x20, 0x2a, 0xff},
		ButtonHot: color.RGBA{0x1e, 0x38, 0x44, 0xff},
		ButtonOn:  color.RGBA{0x1e, 0x44, 0x52, 0xff},
		Field:     color.RGBA{0x06, 0x0c, 0x12, 0xff},
		Focus:     color.RGBA{0x80, 0xff, 0xff, 0xff},
		Selection: color.RGBA{0x24, 0x48, 0x55, 0xff},
		RowH:      rowH,
		Pad:       pad,
		Gap:       gap,
		FieldW:    fieldW,
	}
}

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
	Up, Down       bool
	Home, End      bool
	Enter          bool
	Shift          bool // shift held, to extend the selection with Left/Right/Home/End
	Copy           bool // Ctrl/Cmd+C this frame
	Cut            bool // Ctrl/Cmd+X this frame
	Paste          bool // Ctrl/Cmd+V this frame
	SelectAll      bool // Ctrl/Cmd+A this frame
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
	selAnchor    int            // other end of the selection; == caret means no selection
	clickedField bool           // a field captured the click this frame
	scroll       map[ID]float64 // persisted horizontal scroll per field/list
	vscroll      map[ID]float64 // persisted vertical scroll per text area, in lines

	// Geometry of the last widget, so SameLine can place the next one beside it.
	lastX, lastY, lastW float64

	itemW float64 // fixed width for buttons/toggles (0 = size to text)

	style    Style // colors, metrics and face; defaults to DefaultStyle
	styleSet bool  // whether style has been initialized

	// Panel state, used between BeginPanel and EndPanel. Panels do not nest.
	inPanel              bool
	panelBgIdx           int // draw-command index of the panel background fill
	panelTitleIdx        int // draw-command index of the title-bar fill
	panelX, panelY       float64
	panelMaxX, panelMaxY float64 // running content extent, for auto-sizing
	savedX0              float64 // layout origin saved across the panel

	// Window state. dragWin/dragDX/dragDY persist across frames while dragging.
	curWin         *Window
	dragWin        *Window
	dragDX, dragDY float64

	// Text-selection mouse state. frame counts frames for double-click timing.
	frame          int
	dragField      ID
	lastClickFrame int
	lastClickField ID
}

// Begin starts a frame, laying widgets out from the given top-left position.
func (c *Context) Begin(in Input, x, y float64) {
	c.ensureStyle()
	c.frame++
	c.in = in
	c.x0, c.y0 = x, y
	c.x, c.y = x, y
	c.cmds = c.cmds[:0]
	c.clickedField = false
	c.itemW = 0
	c.inPanel = false
}

// SetItemWidth fixes the width of subsequent buttons and toggles so a column of
// them can share one size (0 sizes each to its text). Reset every frame by Begin.
func (c *Context) SetItemWidth(w float64) {
	c.itemW = w
}

// SetStyle sets the colors, metrics and face used to draw widgets. Start from
// DefaultStyle and override what you need. It persists across frames.
func (c *Context) SetStyle(s Style) {
	c.style = s
	c.styleSet = true
}

// Style returns the Context's current style, initializing it to DefaultStyle if
// it has not been set, so callers can read or tweak individual fields.
func (c *Context) Style() Style {
	c.ensureStyle()
	return c.style
}

// SetFace sets the text face used to measure and draw widget labels, leaving the
// rest of the style untouched. Pass nil to fall back to Ebitengine's built-in
// debug font (a fixed 6x16 cell, ASCII only). A real face also lets non-Latin
// text (e.g. Japanese) render. It persists across frames.
func (c *Context) SetFace(f Face) {
	c.ensureStyle()
	c.style.Face = f
}

// ensureStyle initializes the style to DefaultStyle the first time it is needed,
// so a zero-value Context works without an explicit SetStyle.
func (c *Context) ensureStyle() {
	if !c.styleSet {
		c.style = DefaultStyle()
		c.styleSet = true
	}
}

// End finishes the frame, clearing focus when the click missed every field.
func (c *Context) End() {
	if c.in.MouseClicked && !c.clickedField {
		c.focus = ""
		c.selAnchor = c.caret
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
	c.selAnchor = c.caret
}

// Submitted reports whether Enter was pressed this frame while the field with the
// given id had keyboard focus — handy to run a command typed into a TextField.
// Call it after the field so focus reflects this frame's clicks.
func (c *Context) Submitted(id ID) bool {
	return c.focus == id && c.in.Enter
}

// SameLine places the next widget to the right of the one just drawn, on the
// same row, instead of on a new line below it.
func (c *Context) SameLine() {
	c.x = c.lastX + c.lastW + c.style.Gap
	c.y = c.lastY
}

// Label draws a line of static text.
func (c *Context) Label(s string) {
	rh := c.rowHeight()
	c.textAt(c.x, c.y+(rh-c.fontH())/2, s, c.style.Text)
	c.advance(c.textWidth(s), rh)
}

// Button draws a clickable button and reports whether it was clicked this frame.
func (c *Context) Button(id ID, label string) bool {
	return c.button(label, c.style.Button, c.style.Border)
}

// Toggle is a button that shows an active state, used for tools and on/off
// options; it reports whether it was clicked this frame.
func (c *Context) Toggle(id ID, label string, on bool) bool {
	if on {
		return c.button(label, c.style.ButtonOn, c.style.Focus)
	}
	return c.button(label, c.style.Button, c.style.Border)
}

// swatchSize is the side of a color swatch in logical pixels.
const swatchSize = 18

// Swatch draws a clickable color square, highlighted when selected, and reports
// whether it was clicked this frame.
func (c *Context) Swatch(id ID, col color.RGBA, selected bool) bool {
	w, h := float64(swatchSize), float64(swatchSize)
	hot := within(c.in.MouseX, c.in.MouseY, c.x, c.y, w, h)

	c.fill(c.x, c.y, w, h, col)
	border := c.style.Border
	switch {
	case selected:
		border = c.style.Focus
	case hot:
		border = c.style.Text
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
	w := textW + 2*c.style.Pad
	if c.itemW > w {
		w = c.itemW
	}
	h := c.rowHeight()
	hot := within(c.in.MouseX, c.in.MouseY, c.x, c.y, w, h)
	if hot {
		fill = c.style.ButtonHot
	}
	c.fill(c.x, c.y, w, h, fill)
	c.border(c.x, c.y, w, h, border)
	c.textAt(c.x+(w-textW)/2, c.y+(h-c.fontH())/2, label, c.style.Text)

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
	if c.style.Face == nil {
		ebitenutil.DebugPrintAt(target, cmd.s, int(cmd.x), int(cmd.y))
		return
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(cmd.x, cmd.y)
	op.ColorScale.ScaleWithColor(cmd.col)
	text.Draw(target, cmd.s, c.style.Face, op)
}

// textWidth returns the rendered width of s in logical pixels under the current
// face, or the fixed-cell width when no face is set.
func (c *Context) textWidth(s string) float64 {
	if c.style.Face == nil {
		return float64(len([]rune(s))) * charW
	}
	return text.Advance(s, c.style.Face)
}

// fontH returns the text height used to vertically center labels: the face's line
// height when set, or the fixed debug-font cell height otherwise.
func (c *Context) fontH() float64 {
	if c.style.Face == nil {
		return charH
	}
	m := c.style.Face.Metrics()
	return m.HAscent + m.HDescent
}

// rowHeight is the height of a standard widget row. Without a face it is the
// style's RowH; with one it grows to fit the font's line height.
func (c *Context) rowHeight() float64 {
	if c.style.Face == nil {
		return c.style.RowH
	}
	h := c.fontH() + 2*c.style.Pad
	if h < c.style.RowH {
		h = c.style.RowH
	}
	return h
}

// advance records the just-drawn widget's geometry (so SameLine can use it) and
// moves the layout cursor to the next row.
func (c *Context) advance(w, h float64) {
	c.lastX = c.x
	c.lastY = c.y
	c.lastW = w
	if c.inPanel {
		if r := c.x + w; r > c.panelMaxX {
			c.panelMaxX = r
		}
		if b := c.y + h; b > c.panelMaxY {
			c.panelMaxY = b
		}
	}
	c.newlineH(h)
}

// newlineH advances the layout cursor to the next row, leaving room for a widget
// of the given height.
func (c *Context) newlineH(h float64) {
	c.y += h + c.style.Gap
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
		Up:           repeatKey(ebiten.KeyArrowUp),
		Down:         repeatKey(ebiten.KeyArrowDown),
		Home:         inpututil.IsKeyJustPressed(ebiten.KeyHome),
		End:          inpututil.IsKeyJustPressed(ebiten.KeyEnd),
		Enter:        inpututil.IsKeyJustPressed(ebiten.KeyEnter),
	}
	in.Chars = ebiten.AppendInputChars(nil)
	_, wy := ebiten.Wheel()
	in.WheelY = wy

	in.Shift = ebiten.IsKeyPressed(ebiten.KeyShift)
	if ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta) {
		// A Ctrl/Cmd modifier is held: read shortcuts, and drop text input so the
		// letter is not also typed into a field.
		in.Chars = nil
		in.Copy = inpututil.IsKeyJustPressed(ebiten.KeyC)
		in.Cut = inpututil.IsKeyJustPressed(ebiten.KeyX)
		in.Paste = inpututil.IsKeyJustPressed(ebiten.KeyV)
		in.SelectAll = inpututil.IsKeyJustPressed(ebiten.KeyA)
	}
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
