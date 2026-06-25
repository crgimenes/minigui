package minigui

import "testing"

func driveField(c *Context, s *string, in Input) {
	c.Begin(in, 0, 0)
	c.TextField("f", s)
	c.End()
}

func TestUndoCoalescesTypingRun(t *testing.T) {
	var c Context
	s := ""
	driveField(&c, &s, Input{MouseX: 5, MouseY: 5, MouseClicked: true, MouseDown: true})
	driveField(&c, &s, Input{Chars: []rune("a")})
	driveField(&c, &s, Input{Chars: []rune("b")})
	if s != "ab" {
		t.Fatalf("typed %q, want \"ab\"", s)
	}
	driveField(&c, &s, Input{Undo: true})
	if s != "" {
		t.Fatalf("after one undo %q, want \"\" (whole typing run)", s)
	}
}

func TestUndoSeparateEntriesAcrossNav(t *testing.T) {
	var c Context
	s := ""
	driveField(&c, &s, Input{MouseX: 5, MouseY: 5, MouseClicked: true, MouseDown: true})
	driveField(&c, &s, Input{Chars: []rune("a")})
	driveField(&c, &s, Input{Left: true}) // navigation ends the typing run
	driveField(&c, &s, Input{Chars: []rune("b")})
	if s != "ba" {
		t.Fatalf("got %q, want \"ba\"", s)
	}
	driveField(&c, &s, Input{Undo: true})
	if s != "a" {
		t.Fatalf("undo 1 = %q, want \"a\"", s)
	}
	driveField(&c, &s, Input{Undo: true})
	if s != "" {
		t.Fatalf("undo 2 = %q, want \"\"", s)
	}
}

func TestUndoEmptyStackNoop(t *testing.T) {
	var c Context
	s := "x"
	driveField(&c, &s, Input{MouseX: 5, MouseY: 5, MouseClicked: true, MouseDown: true})
	driveField(&c, &s, Input{Undo: true})
	if s != "x" {
		t.Fatalf("undo with nothing to undo changed text to %q", s)
	}
}

func TestPushPopUndo(t *testing.T) {
	var c Context
	c.pushUndo("f", "a", 1, false)
	c.pushUndo("f", "b", 2, false)
	if snap, ok := c.popUndo("f"); !ok || snap.text != "b" || snap.caret != 2 {
		t.Fatalf("pop = %+v ok=%v, want {b 2} true", snap, ok)
	}
	if snap, ok := c.popUndo("f"); !ok || snap.text != "a" {
		t.Fatalf("pop = %+v ok=%v, want {a 1} true", snap, ok)
	}
	if _, ok := c.popUndo("f"); ok {
		t.Fatal("pop on empty stack should report !ok")
	}
}

func TestPushUndoCoalesceKeepsOneEntry(t *testing.T) {
	var c Context
	c.pushUndo("f", "", 0, true)
	c.pushUndo("f", "a", 1, true) // same run, coalesced away
	if n := len(c.undo["f"]); n != 1 {
		t.Fatalf("coalesced run has %d entries, want 1", n)
	}
}
