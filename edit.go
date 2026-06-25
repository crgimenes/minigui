package minigui

import "github.com/crgimenes/native/clipboard"

// maxUndo bounds the per-field undo history so a long editing session does not
// grow memory without limit.
const maxUndo = 200

// undoSnap is a saved pre-edit state for one field.
type undoSnap struct {
	text  string
	caret int
}

// pushUndo records the pre-edit state for id. A run of plain typing coalesces into
// a single entry, so Ctrl/Cmd+Z undoes the whole run rather than one rune at a time.
func (c *Context) pushUndo(id ID, text string, caret int, coalesce bool) {
	if coalesce && c.undoCoalesce == id {
		return // continue the current typing run; keep the earlier snapshot
	}
	if c.undo == nil {
		c.undo = map[ID][]undoSnap{}
	}
	st := append(c.undo[id], undoSnap{text: text, caret: caret})
	if len(st) > maxUndo {
		st = st[len(st)-maxUndo:]
	}
	c.undo[id] = st
	if coalesce {
		c.undoCoalesce = id
	} else {
		c.undoCoalesce = ""
	}
}

// popUndo removes and returns the last snapshot for id.
func (c *Context) popUndo(id ID) (undoSnap, bool) {
	st := c.undo[id]
	if len(st) == 0 {
		return undoSnap{}, false
	}
	snap := st[len(st)-1]
	c.undo[id] = st[:len(st)-1]
	c.undoCoalesce = ""
	return snap, true
}

// commitEdit applies an edited value to *s, recording undo history. A text change
// pushes the pre-edit state (coalescing typing runs); navigation or a click ends
// the current run so the next typing starts a fresh undo entry.
func (c *Context) commitEdit(id ID, s *string, before string, caretBefore int, ns string, nc, na int) {
	nav := c.in.Left || c.in.Right || c.in.Up || c.in.Down || c.in.Home || c.in.End
	switch {
	case ns != before:
		coalesce := len(c.in.Chars) > 0 && !c.in.Backspace && !c.in.Paste && !c.in.Cut && !c.in.Enter && !nav
		c.pushUndo(id, before, caretBefore, coalesce)
	case nav || c.in.MouseClicked:
		c.undoCoalesce = ""
	}
	*s = ns
	c.caret = nc
	c.selAnchor = na
}

// editFocused runs one frame of editing for the focused field id: undo (Ctrl/Cmd+Z)
// takes priority, otherwise it applies keyboard input via edit, then clipboard
// copy/cut/paste (paste text is filtered by sanitize), recording undo history. It
// returns whether the text changed.
func (c *Context) editFocused(id ID, s *string, edit func(string, int, int, Input) (string, int, int), sanitize func(string) []rune) bool {
	if c.in.Undo {
		snap, ok := c.popUndo(id)
		if !ok {
			return false
		}
		*s = snap.text
		c.caret = clampInt(snap.caret, 0, len([]rune(snap.text)))
		c.selAnchor = c.caret
		return true
	}

	before := *s
	caretBefore := c.caret
	ns, nc, na := edit(*s, c.caret, c.selAnchor, c.in)

	if c.in.Copy || c.in.Cut {
		if lo, hi := selRange(nc, na); lo != hi {
			_ = clipboard.WriteText(string([]rune(ns)[lo:hi]))
			if c.in.Cut {
				ns, nc, na = insertRunes(ns, nc, na, nil)
			}
		}
	}
	if c.in.Paste {
		if txt, err := clipboard.ReadText(); err == nil {
			if ins := sanitize(txt); len(ins) > 0 {
				ns, nc, na = insertRunes(ns, nc, na, ins)
			}
		}
	}

	c.commitEdit(id, s, before, caretBefore, ns, nc, na)
	return ns != before
}
