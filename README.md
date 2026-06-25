# minigui

A tiny immediate-mode GUI toolkit for [Ebitengine](https://ebitengine.org),
meant to be shared across small Ebiten programs that need basic user input
without pulling in a heavy widget framework. Its only dependency is Ebitengine
itself; everything else is the Go standard library.

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
- `List(id, items, *selected) bool` / `ListWithIcons(...)` — scrollable,
  selectable list; the icon variant reserves a square per row for the caller to
  draw into.

Layout helpers: `SameLine`, `SetItemWidth`. Focus helpers: `HasFocus`,
`ClearFocus` (so the host can suspend single-key shortcuts while a field is
being typed into).

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
