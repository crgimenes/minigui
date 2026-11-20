// Package minigui is a tiny immediate-mode GUI toolkit for Ebitengine. Each
// frame the caller builds an Input, runs Begin, calls widgets, then End; widgets
// process input and append draw commands that Render flushes. Keeping the logic
// separate from rendering makes the toolkit testable without a window and keeps a
// consistent look across the host application.
//
// The implementation, and the documentation of each symbol, live in package core.
package minigui

import "github.com/crgimenes/minigui/core"

type (
	Context  = core.Context
	Face     = core.Face
	ID       = core.ID
	Icon     = core.Icon
	IconSlot = core.IconSlot
	Input    = core.Input
	Style    = core.Style
	Window   = core.Window
)

var VGAPalette = core.VGAPalette

func SystemFace(size float64) (Face, error) { return core.SystemFace(size) }

func MustIcon(data []byte) *Icon { return core.MustIcon(data) }

func NewIcon(data []byte) (*Icon, error) { return core.NewIcon(data) }

func InputFromEbiten() Input { return core.InputFromEbiten() }

func DefaultStyle() Style { return core.DefaultStyle() }
