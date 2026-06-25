package minigui

// listRows is how many items a list shows before it scrolls.
const listRows = 6

// IconSlot is the screen-space square reserved for one visible row's icon,
// returned by ListWithIcons. The toolkit only draws fills, borders and text, so
// the caller renders the icon (e.g. an asset thumbnail) into this rect itself.
type IconSlot struct {
	Index int
	X, Y  float64
	Size  float64
}

// List draws a scrollable, selectable list bound to *selected (the chosen index)
// and reports whether the selection changed this frame. Click an item to select
// it; scroll the wheel while hovering to move through a longer list.
func (c *Context) List(id ID, items []string, selected *int) bool {
	changed, _, _ := c.list(id, items, selected, 0)
	return changed
}

// ListWithIcons behaves like List but reserves a square icon column of iconPx on
// the left of each row, insetting the text. It also returns the index clicked
// this frame (-1 if none, even when it equals the current selection, so callers
// can detect double-clicks) and the icon rect of each visible row so the caller
// can draw into it after Render; the slices/values are valid until the next frame.
func (c *Context) ListWithIcons(id ID, items []string, selected *int, iconPx float64) (changed bool, clicked int, icons []IconSlot) {
	return c.list(id, items, selected, iconPx)
}

func (c *Context) list(id ID, items []string, selected *int, iconPx float64) (changed bool, clicked int, icons []IconSlot) {
	clicked = -1
	visible := max(min(len(items), listRows),
		// keep a minimal box even when the list is empty
		1)
	w := float64(fieldW)
	rh := c.rowHeight()
	if iconPx+2 > rh {
		rh = iconPx + 2 // grow the row so a tall icon still fits
	}
	h := float64(visible) * rh

	maxOff := max(len(items)-listRows, 0)
	pos := c.scroll[id]
	if within(c.in.MouseX, c.in.MouseY, c.x, c.y, w, h) {
		pos -= c.in.WheelY // wheel down (negative dy) scrolls toward later items
	}
	if pos < 0 {
		pos = 0
	}
	if pos > float64(maxOff) {
		pos = float64(maxOff)
	}
	c.setScroll(id, pos)
	off := int(pos)

	c.fill(c.x, c.y, w, h, colField)

	textX := c.x + pad
	if iconPx > 0 {
		textX = c.x + pad + iconPx + pad
	}

	for i := 0; i < visible; i++ {
		idx := off + i
		if idx >= len(items) {
			break
		}
		ry := c.y + float64(i)*rh
		itemHot := within(c.in.MouseX, c.in.MouseY, c.x, ry, w, rh)
		switch {
		case idx == *selected:
			c.fill(c.x, ry, w, rh, colBtnHot)
		case itemHot:
			c.fill(c.x, ry, w, rh, colBtn)
		}
		if itemHot && c.in.MouseClicked {
			clicked = idx
			if *selected != idx {
				*selected = idx
				changed = true
			}
		}
		if iconPx > 0 {
			icons = append(icons, IconSlot{Index: idx, X: c.x + pad, Y: ry + (rh-iconPx)/2, Size: iconPx})
		}
		c.textAt(textX, ry+(rh-c.fontH())/2, items[idx], colText)
	}
	c.border(c.x, c.y, w, h, colBorder)

	// Scroll thumb on the right edge when the list overflows.
	if len(items) > listRows {
		thumbH := h * float64(listRows) / float64(len(items))
		if thumbH < 8 {
			thumbH = 8
		}
		thumbY := c.y
		if maxOff > 0 {
			thumbY = c.y + (h-thumbH)*pos/float64(maxOff)
		}
		c.fill(c.x+w-3, thumbY, 3, thumbH, colBorder)
	}

	c.advance(w, h)
	return changed, clicked, icons
}
