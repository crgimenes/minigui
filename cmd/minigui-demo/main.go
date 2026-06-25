// Command minigui-demo is a sandbox for the minigui package: a panel with a
// label, a button and an editable text field, so the immediate-mode toolkit can
// be exercised on its own. It loads a system font so non-Latin text (Japanese
// here) renders, falling back to the built-in debug font when none is found.
package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/crgimenes/minigui"
)

type demo struct {
	gui   minigui.Context
	count int
	name  string
	notes string
	items []string
	sel   int
}

func (d *demo) Update() error {
	in := minigui.InputFromEbiten()
	d.gui.Begin(in, 40, 40)
	d.gui.Label("minigui demo — 日本語 / CJK")
	if d.gui.Button("inc", fmt.Sprintf("count: %d", d.count)) {
		d.count++
	}
	d.gui.TextField("name", &d.name)
	d.gui.List("items", d.items, &d.sel)
	d.gui.Label(fmt.Sprintf("selected: %s", d.items[d.sel]))
	d.gui.Label("notes (multi-line):")
	d.gui.TextArea("notes", &d.notes, 4)
	d.gui.End()
	return nil
}

func (d *demo) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x06, 0x08, 0x0c, 0xff})
	d.gui.Render(screen)
}

func (d *demo) Layout(int, int) (int, int) {
	return 520, 560
}

func main() {
	items := make([]string, 12)
	for i := range items {
		items[i] = fmt.Sprintf("item %02d", i)
	}

	d := &demo{name: "world", notes: "type here\nmultiple lines\nshift+arrows select", items: items}
	if face, err := minigui.SystemFace(16); err != nil {
		log.Printf("minigui-demo: %v; using debug font", err)
	} else {
		d.gui.SetFace(face)
	}

	ebiten.SetWindowSize(520, 560)
	ebiten.SetWindowTitle("minigui demo")
	if err := ebiten.RunGame(d); err != nil {
		panic(err)
	}
}
