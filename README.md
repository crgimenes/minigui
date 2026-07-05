# minigui

A tiny immediate-mode GUI toolkit for [Ebitengine](https://ebitengine.org),
meant to be shared across small Ebiten programs that need basic user input
without pulling in a heavy widget framework. Its dependencies are Ebitengine
itself and [`native`](https://github.com/crgimenes/native) (cgo-free system
clipboard for copy/cut/paste); everything else is the Go standard library.

> Status: early but usable — pluggable text faces (see [Fonts](#fonts)) and
> per-app styling (see [Theming](#theming)).

## Model

Immediate mode: every frame you build an `Input`, call `Begin`, run the widgets,
then `End`; widgets read the input and append draw commands that `Render` flushes
to an `*ebiten.Image`. Input is a plain struct, so the widget logic is testable
without a window.

```go
type game struct {
	gui   minigui.Context
	name  string
	count int
}

func (g *game) Update() error {
	g.gui.Begin(minigui.InputFromEbiten(), 20, 20)
	g.gui.Label("hello")
	if g.gui.Button("inc", fmt.Sprintf("count: %d", g.count)) {
		g.count++
	}
	g.gui.TextField("name", &g.name)
	g.gui.End()
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	g.gui.Render(screen)
}
```

## Widgets

- `Label(text)` — static text.
- `Button(id, label) bool` — reports a click this frame.
- `Toggle(id, label, on) bool` — button with an active state.
- `Swatch(id, color, selected) bool` — clickable color square.
- `TextField(id, *string) bool` — editable single-line field (focus, caret,
  horizontal scroll); reports whether the text changed.
- `TextArea(id, *string, rows) bool` — multi-line editor (Enter inserts a line,
  Up/Down move between lines, line-relative Home/End, vertical scroll).
- `List(id, items, *selected) bool` / `ListWithIcons(...)` — scrollable,
  selectable list; the icon variant reserves a square per row for the caller to
  draw into.

Text editing (TextField and TextArea) comes with selection (Shift+arrows, mouse
drag, double-click word select, Ctrl/Cmd+A), system-clipboard copy/cut/paste
(Ctrl/Cmd+C/X/V, via `native/clipboard`), and undo (Ctrl/Cmd+Z, with typing
runs coalesced).

Layout helpers: `SameLine`, `SetItemWidth`. Focus helpers: `HasFocus`,
`ClearFocus` (so the host can suspend single-key shortcuts while a field is
being typed into), `Focus(id)` to hand focus to a widget, and `Submitted(id)`
to detect Enter in a field.

## Fonts

By default a `Context` renders labels with Ebitengine's built-in debug font (a
fixed 6×16 cell, ASCII only). Call `SetFace` to use any
[`text/v2`](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/text/v2) face —
then text is measured and drawn through it (so widths, the caret and row heights
follow the font), non-Latin scripts such as Japanese render, and label colors
are honored.

`SystemFace(size)` is a convenience that loads a CJK-capable operating-system
font (it tries a small per-OS list of well-known paths, no font bundled, no
external dependency) and returns a face for `SetFace`. It returns an error when
none is found, so you can fall back to the debug font:

```go
if face, err := minigui.SystemFace(16); err == nil {
	gui.SetFace(face)
}
```

To bundle your own font instead, build a face from embedded bytes with
`text.NewGoTextFaceSource` and pass it to `SetFace`.

## Theming

Colors and layout metrics live in a `Style`. A zero-value `Context` uses
`DefaultStyle` (a dark, cyan-tinted scheme); override what you want and apply it
with `SetStyle`:

```go
s := minigui.DefaultStyle()
s.Button = color.RGBA{0x20, 0x14, 0x28, 0xff}
s.Focus = color.RGBA{0xff, 0x99, 0x33, 0xff}
s.RowH = 28
s.Face = face // optional; same as SetFace
gui.SetStyle(s)
```

`SetFace` is a shortcut that sets only `Style.Face`; `Style()` returns the
current style for reading or tweaking a single field.

`VGAPalette` is the classic 16-color IBM VGA text palette as hex strings —
handy as a ready-made swatch set for `Swatch` rows in retro-styled tools.

## Panels

`BeginPanel(title, x, y)` / `EndPanel()` frame the widgets between them in a
titled background box (a small simulated window), auto-sized to the content.
`EndPanel` returns the panel's screen rectangle, handy for hit-testing the whole
panel — e.g. to keep it interactive in a click-through overlay.

```go
gui.Begin(in, 0, 0)
gui.BeginPanel("Tools", 8, 8)
gui.Toggle("draw", "Draw", drawing)
rect := gui.EndPanel()
gui.End()
```

## Windows

A `Window` is a panel with a life of its own: draggable by the title bar and
closable by a close box. The caller owns the state (`X`, `Y`, `Open` persist
across frames); minigui moves `X`/`Y` during a drag and clears `Open` on close
(`NoClose` hides the close box, e.g. for an always-present toolbar).
`Dragging()` reports an in-progress drag, so a click-through host can keep
grabbing the mouse until the drag ends.

```go
var win = minigui.Window{Title: "Tools", X: 8, Y: 8, Open: true}

if gui.BeginWindow(&win) {
	gui.Toggle("draw", "Draw", drawing)
	gui.EndWindow()
}
```

## Install

```sh
go get github.com/crgimenes/minigui
```

## Demo

```sh
go run github.com/crgimenes/minigui/cmd/minigui-demo
```

## License

See [`LICENSE`](LICENSE).

---

## More of my projects

- [native](https://github.com/crgimenes/native): cgo-free Go bindings for OS APIs: clipboard, mmap, keep-awake, and friends.
- [NeoFrame](https://github.com/crgimenes/NeoFrame): draw over your screen; a transparent overlay for demos and classes.
- [kutta](https://github.com/crgimenes/kutta): a 2D wind tunnel; watch air misbehave around an airfoil.
- [glaze](https://github.com/crgimenes/glaze): WebView desktop apps in Go, cgo-free.

More at [github.com/crgimenes](https://github.com/crgimenes) and [crg.eti.br](https://crg.eti.br).
