package minigui

import "image"

// panelStart reserves the panel chrome (background + title bar + title) at (x, y)
// and positions the layout cursor inside, below the title bar. It is shared by
// BeginPanel and BeginWindow; the chrome size is patched by panelFinish.
func (c *Context) panelStart(title string, x, y float64) {
	c.ensureStyle()
	c.inPanel = true
	c.panelX, c.panelY = x, y

	// Reserve the background and title-bar fills now, before the widgets, so they
	// render behind them; their final size is patched in panelFinish.
	c.panelBgIdx = len(c.cmds)
	c.fill(x, y, 0, 0, c.style.Field)

	titleH := c.rowHeight()
	c.panelTitleIdx = len(c.cmds)
	c.fill(x, y, 0, titleH, c.style.Button)
	c.textAt(x+c.style.Pad, y+(titleH-c.fontH())/2, title, c.style.Text)

	c.savedX0 = c.x0
	c.x0 = x + c.style.Pad
	c.x = c.x0
	c.y = y + titleH + c.style.Pad

	c.panelMaxX = x + c.style.Pad + c.textWidth(title)
	c.panelMaxY = c.y
}

// panelFinish sizes the chrome to the content, draws the border and returns the
// panel's screen rectangle plus its title-bar height. Shared by EndPanel and
// EndWindow.
func (c *Context) panelFinish() (image.Rectangle, float64) {
	w := c.panelMaxX - c.panelX + c.style.Pad
	h := c.panelMaxY - c.panelY + c.style.Pad

	c.cmds[c.panelBgIdx].w = w
	c.cmds[c.panelBgIdx].h = h
	c.cmds[c.panelTitleIdx].w = w
	c.border(c.panelX, c.panelY, w, h, c.style.Border)

	c.inPanel = false
	c.x0 = c.savedX0
	c.x = c.x0

	titleH := c.cmds[c.panelTitleIdx].h
	return image.Rect(int(c.panelX), int(c.panelY), int(c.panelX+w), int(c.panelY+h)), titleH
}

// BeginPanel starts a titled panel at (x, y): a fixed background box with a title
// bar that frames the widgets drawn until EndPanel, like a small simulated window.
// The box is sized automatically to its content. Panels do not nest. For a movable,
// closable window backed by caller state, use BeginWindow.
func (c *Context) BeginPanel(title string, x, y float64) {
	c.panelStart(title, x, y)
}

// EndPanel finishes the panel and returns its screen rectangle, useful for
// hit-testing the whole panel (e.g. to keep it interactive in a click-through
// overlay).
func (c *Context) EndPanel() image.Rectangle {
	rect, _ := c.panelFinish()
	return rect
}

// Window is the persistent state of a draggable, closable window. The caller owns
// it: X/Y and Open survive across frames. minigui moves X/Y while the title bar is
// dragged and clears Open when the close box is clicked (unless NoClose is set).
type Window struct {
	Title   string
	X, Y    float64
	Open    bool
	NoClose bool // hide the close box, e.g. for an always-present toolbar
}

// BeginWindow starts the window w. It returns false when w is closed (w.Open is
// false): skip its contents and do not call EndWindow. When it returns true, draw
// widgets and call EndWindow.
func (c *Context) BeginWindow(w *Window) bool {
	if w == nil || !w.Open {
		return false
	}
	c.curWin = w
	c.panelStart(w.Title, w.X, w.Y)
	return true
}

// EndWindow finishes the current window: it draws the close box, handles dragging
// the title bar and clicking the close box (mutating the caller's Window), and
// returns the window's screen rectangle.
func (c *Context) EndWindow() image.Rectangle {
	w := c.curWin
	rect, titleH := c.panelFinish()
	c.curWin = nil

	right := float64(rect.Max.X)
	top := float64(rect.Min.Y)
	closeX := right - titleH

	// Close box at the top-right of the title bar.
	if !w.NoClose {
		hot := within(c.in.MouseX, c.in.MouseY, closeX, top, titleH, titleH)
		if hot {
			c.fill(closeX, top, titleH, titleH, c.style.ButtonHot)
		}
		c.textAt(closeX+(titleH-c.textWidth("x"))/2, top+(titleH-c.fontH())/2, "x", c.style.Text)
		if hot && c.in.MouseClicked {
			w.Open = false
		}
	}

	// Drag the title bar (excluding the close box) to move the window. The grab is
	// tracked by window pointer so it continues even if the cursor outruns the bar.
	dragW := right - float64(rect.Min.X)
	if !w.NoClose {
		dragW = closeX - float64(rect.Min.X)
	}
	if c.dragWin == nil && c.in.MouseClicked && within(c.in.MouseX, c.in.MouseY, float64(rect.Min.X), top, dragW, titleH) {
		c.dragWin = w
		c.dragDX = c.in.MouseX - w.X
		c.dragDY = c.in.MouseY - w.Y
	}
	if c.dragWin == w {
		if c.in.MouseDown {
			w.X = c.in.MouseX - c.dragDX
			w.Y = c.in.MouseY - c.dragDY
		} else {
			c.dragWin = nil
		}
	}

	return rect
}

// Dragging reports whether a window is currently being dragged, so a click-through
// host can keep grabbing the mouse until the drag ends, even if the cursor briefly
// outruns the title bar.
func (c *Context) Dragging() bool {
	return c.dragWin != nil
}
