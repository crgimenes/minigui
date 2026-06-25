package minigui

import (
	"image"

	"github.com/crgimenes/native/clipboard"
)

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// selRange returns the selection bounds [lo, hi) for a caret/anchor pair; lo == hi
// means there is no selection.
func selRange(caret, anchor int) (lo, hi int) {
	if caret <= anchor {
		return caret, anchor
	}
	return anchor, caret
}

// insertRunes replaces the selection [lo, hi) (or, with no selection, the caret
// position) with ins, and returns the new text with a collapsed caret/anchor just
// after the inserted runes. Passing a nil ins deletes the selection.
func insertRunes(s string, caret, anchor int, ins []rune) (string, int, int) {
	r := []rune(s)
	lo, hi := selRange(clampInt(caret, 0, len(r)), clampInt(anchor, 0, len(r)))
	out := make([]rune, 0, len(r)-(hi-lo)+len(ins))
	out = append(out, r[:lo]...)
	out = append(out, ins...)
	out = append(out, r[hi:]...)
	caret = lo + len(ins)
	return string(out), caret, caret
}

// inlineRunes drops control characters (including newlines and tabs) so pasted
// text stays on a single line.
func inlineRunes(s string) []rune {
	out := make([]rune, 0, len(s))
	for _, ch := range s {
		if ch < 0x20 || ch == 0x7f {
			continue
		}
		out = append(out, ch)
	}
	return out
}

// editText applies one frame of keyboard input to a single-line buffer with a
// selection, returning the new text, caret and anchor. It is pure (no clipboard)
// so the editing logic is testable without a window; Shift extends the selection,
// SelectAll selects everything, and typing or Backspace replaces the selection.
func editText(s string, caret, anchor int, in Input) (string, int, int) {
	r := []rune(s)
	n := len(r)
	caret = clampInt(caret, 0, n)
	anchor = clampInt(anchor, 0, n)

	if in.SelectAll {
		return s, n, 0
	}

	lo, hi := selRange(caret, anchor)

	if len(in.Chars) > 0 {
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
		} else if caret < len(r) {
			caret++
		}
		moved = true
	}
	if in.Home {
		caret = 0
		moved = true
	}
	if in.End {
		caret = len(r)
		moved = true
	}
	if moved && !in.Shift {
		anchor = caret
	}
	return string(r), caret, anchor
}

// TextField draws an editable single-line field bound to *s and reports whether
// the text changed this frame. Clicking it gives it focus; while focused it edits
// the text, supports a selection (Shift + arrows, Ctrl/Cmd+A) and clipboard
// copy/cut/paste (Ctrl/Cmd+C/X/V) via the system clipboard.
func (c *Context) TextField(id ID, s *string) bool {
	w, h := c.style.FieldW, c.rowHeight()
	hot := within(c.in.MouseX, c.in.MouseY, c.x, c.y, w, h)
	if hot && c.in.MouseClicked {
		c.focus = id
		c.caret = len([]rune(*s))
		c.selAnchor = c.caret
		c.clickedField = true
	}

	focused := c.focus == id
	changed := false
	if focused {
		before := *s
		ns, nc, na := editText(*s, c.caret, c.selAnchor, c.in)

		if c.in.Copy || c.in.Cut {
			if lo, hi := selRange(nc, na); lo != hi {
				sel := string([]rune(ns)[lo:hi])
				_ = clipboard.WriteText(sel)
				if c.in.Cut {
					ns, nc, na = insertRunes(ns, nc, na, nil)
				}
			}
		}
		if c.in.Paste {
			if txt, err := clipboard.ReadText(); err == nil {
				if ins := inlineRunes(txt); len(ins) > 0 {
					ns, nc, na = insertRunes(ns, nc, na, ins)
				}
			}
		}

		*s = ns
		c.caret = nc
		c.selAnchor = na
		changed = ns != before
	}

	// Horizontal scroll: slide the view so the caret stays inside the box. The
	// caret position is the width of the text before it, so it tracks proportional
	// fonts, and the glyphs are clipped so nothing spills past the border.
	inner := w - 2*c.style.Pad
	runes := []rune(*s)
	caret := clampInt(c.caret, 0, len(runes))
	caretPx := c.textWidth(string(runes[:caret]))
	textW := c.textWidth(*s)

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
	if maxX := textW - inner; xoff > maxX {
		xoff = maxX
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

	textX := c.x + c.style.Pad - xoff

	// Selection highlight behind the text, clamped to the field interior.
	if focused {
		if lo, hi := selRange(caret, clampInt(c.selAnchor, 0, len(runes))); lo != hi {
			x0 := textX + c.textWidth(string(runes[:lo]))
			x1 := textX + c.textWidth(string(runes[:hi]))
			left, right := c.x+1, c.x+w-1
			x0 = clampF(x0, left, right)
			x1 = clampF(x1, left, right)
			if x1 > x0 {
				c.fill(x0, c.y+2, x1-x0, h-4, c.style.Selection)
			}
		}
	}

	clip := image.Rect(int(c.x)+1, int(c.y)+1, int(c.x+w)-1, int(c.y+h)-1)
	c.textClip(textX, c.y+(h-c.fontH())/2, *s, c.style.Text, clip)
	if focused {
		c.fill(textX+caretPx, c.y+3, 1, h-6, c.style.Focus)
	}

	c.advance(w, h)
	return changed
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
