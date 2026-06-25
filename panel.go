package minigui

import "image"

// BeginPanel starts a titled panel at (x, y): a background box with a title bar
// that frames the widgets drawn until EndPanel, like a small simulated window.
// The box is sized automatically to its content. Panels do not nest.
func (c *Context) BeginPanel(title string, x, y float64) {
	c.ensureStyle()
	c.inPanel = true
	c.panelX, c.panelY = x, y

	// Reserve the background and title-bar fills now, before the widgets, so they
	// render behind them; their final width and height are patched in EndPanel,
	// once the content extent is known.
	c.panelBgIdx = len(c.cmds)
	c.fill(x, y, 0, 0, c.style.Field)

	titleH := c.rowHeight()
	c.panelTitleIdx = len(c.cmds)
	c.fill(x, y, 0, titleH, c.style.Button)
	c.textAt(x+c.style.Pad, y+(titleH-c.fontH())/2, title, c.style.Text)

	// Lay widgets out inside the panel, below the title bar.
	c.savedX0 = c.x0
	c.x0 = x + c.style.Pad
	c.x = c.x0
	c.y = y + titleH + c.style.Pad

	// Seed the content extent so the panel is at least as wide as its title.
	c.panelMaxX = x + c.style.Pad + c.textWidth(title)
	c.panelMaxY = c.y
}

// EndPanel finishes the panel, sizing its background to the content, and returns
// the panel's screen rectangle so the caller can hit-test the whole panel (e.g.
// to keep it interactive in a click-through overlay).
func (c *Context) EndPanel() image.Rectangle {
	w := c.panelMaxX - c.panelX + c.style.Pad
	h := c.panelMaxY - c.panelY + c.style.Pad

	c.cmds[c.panelBgIdx].w = w
	c.cmds[c.panelBgIdx].h = h
	c.cmds[c.panelTitleIdx].w = w
	c.border(c.panelX, c.panelY, w, h, c.style.Border)

	c.inPanel = false
	c.x0 = c.savedX0
	c.x = c.x0

	return image.Rect(int(c.panelX), int(c.panelY), int(c.panelX+w), int(c.panelY+h))
}
