package minigui

import "image"

// lineStart returns the index of the first rune of the line containing pos (the
// rune just after the previous newline, or 0).
func lineStart(r []rune, pos int) int {
	for i := pos - 1; i >= 0; i-- {
		if r[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

// lineEnd returns the index of the newline ending the line containing pos, or
// len(r) for the last line.
func lineEnd(r []rune, pos int) int {
	for i := pos; i < len(r); i++ {
		if r[i] == '\n' {
			return i
		}
	}
	return len(r)
}

// caretUp moves pos to the previous line, keeping the column when possible.
func caretUp(r []rune, pos int) int {
	ls := lineStart(r, pos)
	if ls == 0 {
		return pos // already on the first line
	}
	col := pos - ls
	prevStart := lineStart(r, ls-1)
	prevLen := ls - 1 - prevStart
	return prevStart + min(col, prevLen)
}

// caretDown moves pos to the next line, keeping the column when possible.
func caretDown(r []rune, pos int) int {
	le := lineEnd(r, pos)
	if le == len(r) {
		return pos // already on the last line
	}
	col := pos - lineStart(r, pos)
	nextStart := le + 1
	nextLen := lineEnd(r, nextStart) - nextStart
	return nextStart + min(col, nextLen)
}

// multilineRunes drops control characters but keeps newlines, so pasted text
// keeps its line breaks in a text area.
func multilineRunes(s string) []rune {
	out := make([]rune, 0, len(s))
	for _, ch := range s {
		if ch == '\n' {
			out = append(out, ch)
			continue
		}
		if ch < 0x20 || ch == 0x7f {
			continue
		}
		out = append(out, ch)
	}
	return out
}

// editTextArea applies one frame of keyboard input to a multi-line buffer with a
// selection, returning the new text, caret and anchor. Like editText but Enter
// inserts a newline, Up/Down move between lines, and Home/End are line-relative.
func editTextArea(s string, caret, anchor int, in Input) (string, int, int) {
	r := []rune(s)
	n := len(r)
	caret = clampInt(caret, 0, n)
	anchor = clampInt(anchor, 0, n)

	if in.SelectAll {
		return s, n, 0
	}

	lo, hi := selRange(caret, anchor)

	if len(in.Chars) > 0 || in.Enter {
		if lo != hi {
			r = append(r[:lo], r[hi:]...)
			caret = lo
		}
		for _, ch := range in.Chars {
			if ch < 0x20 || ch == 0x7f {
				continue
			}
			r = append(r[:caret], append([]rune{ch}, r[caret:]...)...)
			caret++
		}
		if in.Enter {
			r = append(r[:caret], append([]rune{'\n'}, r[caret:]...)...)
			caret++
		}
		return string(r), caret, caret
	}

	if in.Backspace {
		if lo != hi {
			r = append(r[:lo], r[hi:]...)
			caret = lo
		} else if caret > 0 {
			r = append(r[:caret-1], r[caret:]...)
			caret--
		}
		return string(r), caret, caret
	}

	moved := false
	if in.Left {
		if !in.Shift && lo != hi {
			caret = lo
		} else if caret > 0 {
			caret--
		}
		moved = true
	}
	if in.Right {
		if !in.Shift && lo != hi {
			caret = hi
		} else if caret < n {
			caret++
		}
		moved = true
	}
	if in.Up {
		caret = caretUp(r, caret)
		moved = true
	}
	if in.Down {
		caret = caretDown(r, caret)
		moved = true
	}
	if in.Home {
		caret = lineStart(r, caret)
		moved = true
	}
	if in.End {
		caret = lineEnd(r, caret)
		moved = true
	}
	if moved && !in.Shift {
		anchor = caret
	}
	return string(r), caret, anchor
}

// setVScroll stores a text area's vertical scroll (top line), lazily creating the
// map.
func (c *Context) setVScroll(id ID, v float64) {
	if c.vscroll == nil {
		c.vscroll = map[ID]float64{}
	}
	c.vscroll[id] = v
}

// TextArea draws an editable multi-line field bound to *s, showing rows lines, and
// reports whether the text changed this frame. It supports the same selection and
// clipboard shortcuts as TextField, plus Enter for a newline and Up/Down to move
// between lines; long lines scroll horizontally (no wrap).
func (c *Context) TextArea(id ID, s *string, rows int) bool {
	if rows < 1 {
		rows = 1
	}
	lineH := c.fontH() + 2
	w := c.style.FieldW
	h := float64(rows)*lineH + 2*c.style.Pad

	runes := []rune(*s)
	lineStarts := []int{0}
	for i, ch := range runes {
		if ch == '\n' {
			lineStarts = append(lineStarts, i+1)
		}
	}
	numLines := len(lineStarts)
	contentEnd := func(li int) int {
		if li+1 < numLines {
			return lineStarts[li+1] - 1
		}
		return len(runes)
	}
	lineOf := func(idx int) int {
		li := 0
		for i := 0; i < numLines; i++ {
			if lineStarts[i] <= idx {
				li = i
			} else {
				break
			}
		}
		return li
	}

	hot := within(c.in.MouseX, c.in.MouseY, c.x, c.y, w, h)
	textLeft := c.x + c.style.Pad
	top := int(c.vscroll[id])

	// indexAt maps a mouse position to a flat rune index, using the current top
	// line and horizontal offset.
	indexAt := func(mx, my, xoff float64) int {
		vis := int((my - (c.y + c.style.Pad)) / lineH)
		li := clampInt(top+vis, 0, numLines-1)
		ls, ce := lineStarts[li], contentEnd(li)
		col := c.caretAtX(string(runes[ls:ce]), mx-textLeft+xoff)
		return ls + col
	}

	// Mouse: click positions the caret, Shift+click and drag extend the selection,
	// double-click selects the word.
	prevXoff := c.scroll[id]
	switch {
	case hot && c.in.MouseClicked:
		idx := indexAt(c.in.MouseX, c.in.MouseY, prevXoff)
		c.focus = id
		c.clickedField = true
		c.dragField = id
		double := c.lastClickField == id && c.frame-c.lastClickFrame <= doubleClickFrames
		switch {
		case double:
			c.selAnchor, c.caret = wordBounds(*s, idx)
		case c.in.Shift:
			c.caret = idx
		default:
			c.caret, c.selAnchor = idx, idx
		}
		c.lastClickField = id
		c.lastClickFrame = c.frame
	case c.dragField == id && c.in.MouseDown:
		c.caret = indexAt(c.in.MouseX, c.in.MouseY, prevXoff)
	}
	if !c.in.MouseDown {
		c.dragField = ""
	}

	focused := c.focus == id
	changed := false
	if focused {
		changed = c.editFocused(id, s, editTextArea, multilineRunes)

		// Recompute the line layout if the text changed under us.
		runes = []rune(*s)
		lineStarts = lineStarts[:1]
		for i, ch := range runes {
			if ch == '\n' {
				lineStarts = append(lineStarts, i+1)
			}
		}
		numLines = len(lineStarts)
	}

	caret := clampInt(c.caret, 0, len(runes))
	anchor := clampInt(c.selAnchor, 0, len(runes))
	caretLine := lineOf(caret)
	caretPx := c.textWidth(string(runes[lineStarts[caretLine]:caret]))

	// Vertical scroll: keep the caret line on screen.
	if caretLine < top {
		top = caretLine
	}
	if caretLine >= top+rows {
		top = caretLine - rows + 1
	}
	maxTop := max(numLines-rows, 0)
	top = clampInt(top, 0, maxTop)
	c.setVScroll(id, float64(top))

	// Horizontal scroll: keep the caret column in view.
	inner := w - 2*c.style.Pad
	xoff := c.scroll[id]
	if focused {
		if caretPx-xoff < 0 {
			xoff = caretPx
		}
		if caretPx-xoff > inner {
			xoff = caretPx - inner
		}
	} else {
		xoff = 0
	}
	if xoff < 0 {
		xoff = 0
	}
	c.setScroll(id, xoff)

	border := c.style.Border
	if focused {
		border = c.style.Focus
	}
	c.fill(c.x, c.y, w, h, c.style.Field)
	c.border(c.x, c.y, w, h, border)

	textX0 := textLeft - xoff
	clip := image.Rect(int(c.x)+1, int(c.y)+1, int(c.x+w)-1, int(c.y+h)-1)
	left, right := c.x+1, c.x+w-1
	lo, hi := selRange(caret, anchor)

	for vis := 0; vis < rows; vis++ {
		li := top + vis
		if li >= numLines {
			break
		}
		ls, ce := lineStarts[li], contentEnd(li)
		lineY := c.y + c.style.Pad + float64(vis)*lineH

		if focused && lo != hi {
			ss := max(lo, ls)
			se := min(hi, ce)
			selNewline := li+1 < numLines && lo <= ce && ce < hi
			if ss < se || selNewline {
				x0 := textX0 + c.textWidth(string(runes[ls:ss]))
				x1 := textX0 + c.textWidth(string(runes[ls:se]))
				if selNewline {
					x1 = right
				}
				x0 = clampF(x0, left, right)
				x1 = clampF(x1, left, right)
				if x1 > x0 {
					c.fill(x0, lineY, x1-x0, lineH, c.style.Selection)
				}
			}
		}

		c.textClip(textX0, lineY+(lineH-c.fontH())/2, string(runes[ls:ce]), c.style.Text, clip)
	}

	if focused && caretLine >= top && caretLine < top+rows {
		caretY := c.y + c.style.Pad + float64(caretLine-top)*lineH
		c.fill(textX0+caretPx, caretY+(lineH-c.fontH())/2, 1, c.fontH(), c.style.Focus)
	}

	c.advance(w, h)
	return changed
}
