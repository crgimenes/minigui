package minigui

import "image"

// editText applies one frame of input to a single-line text buffer and returns
// the new text and caret position. It is pure so the editing logic is testable
// without a window. Control characters are ignored; the caret is clamped.
func editText(s string, caret int, in Input) (string, int) {
	r := []rune(s)
	if caret < 0 {
		caret = 0
	}
	if caret > len(r) {
		caret = len(r)
	}

	for _, ch := range in.Chars {
		if ch < 0x20 || ch == 0x7f {
			continue
		}
		r = append(r[:caret], append([]rune{ch}, r[caret:]...)...)
		caret++
	}
	if in.Backspace && caret > 0 {
		r = append(r[:caret-1], r[caret:]...)
		caret--
	}
	if in.Left && caret > 0 {
		caret--
	}
	if in.Right && caret < len(r) {
		caret++
	}
	if in.Home {
		caret = 0
	}
	if in.End {
		caret = len(r)
	}
	return string(r), caret
}

// TextField draws an editable single-line field bound to *s and reports whether
// the text changed this frame. Clicking it gives it focus; it edits only while
// focused.
func (c *Context) TextField(id ID, s *string) bool {
	w, h := float64(fieldW), c.rowHeight()
	hot := within(c.in.MouseX, c.in.MouseY, c.x, c.y, w, h)
	if hot && c.in.MouseClicked {
		c.focus = id
		c.caret = len([]rune(*s))
		c.clickedField = true
	}

	focused := c.focus == id
	changed := false
	if focused {
		ns, nc := editText(*s, c.caret, c.in)
		if ns != *s {
			*s = ns
			changed = true
		}
		c.caret = nc
	}

	// Horizontal scroll: keep the whole text but slide the view so the caret stays
	// inside the box (a single-line field never limits how much you can type). The
	// glyphs are clipped to the box so nothing spills past the border. The caret
	// position is the width of the text before it, so it tracks proportional fonts.
	inner := w - 2*pad
	textW := c.textWidth(*s)
	xoff := c.scroll[id]
	caretPx := 0.0
	if focused {
		runes := []rune(*s)
		caret := min(max(c.caret, 0), len(runes))
		caretPx = c.textWidth(string(runes[:caret]))
		if caretPx-xoff < 0 {
			xoff = caretPx
		}
		if caretPx-xoff > inner {
			xoff = caretPx - inner
		}
	} else {
		xoff = 0
	}
	if maxX := textW - inner; xoff > maxX {
		xoff = maxX
	}
	if xoff < 0 {
		xoff = 0
	}
	c.setScroll(id, xoff)

	border := colBorder
	if focused {
		border = colFocus
	}
	c.fill(c.x, c.y, w, h, colField)
	c.border(c.x, c.y, w, h, border)

	clip := image.Rect(int(c.x)+1, int(c.y)+1, int(c.x+w)-1, int(c.y+h)-1)
	c.textClip(c.x+pad-xoff, c.y+(h-c.fontH())/2, *s, colText, clip)
	if focused {
		caretX := c.x + pad + caretPx - xoff
		c.fill(caretX, c.y+3, 1, h-6, colFocus)
	}

	c.advance(w, h)
	return changed
}
