package minigui

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// SystemFace loads a text face from a well-known operating-system font at the
// given size in pixels, ready to pass to (*Context).SetFace. It tries a small,
// per-OS list of font paths (CJK-capable ones first, so Japanese and other
// non-Latin text render) and returns the first that loads. When none load it
// returns an error, so the caller can fall back to the built-in debug font by
// leaving the Context's face unset.
//
// It depends only on the standard library and Ebitengine; it does no font-config
// discovery, so the lists are best-effort and can miss unusual installations.
func SystemFace(size float64) (Face, error) {
	candidates := systemFontCandidates()
	for _, path := range candidates {
		src, err := loadFaceSource(path)
		if err != nil {
			continue
		}
		return &text.GoTextFace{Source: src, Size: size}, nil
	}
	return nil, fmt.Errorf("minigui: no usable system font found (%d candidates tried)", len(candidates))
}

// loadFaceSource reads a font file into a face source. TrueType/OpenType
// collections (.ttc/.otc) hold several fonts; the first is used.
func loadFaceSource(path string) (*text.GoTextFaceSource, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is from a fixed per-OS allowlist, never user input
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".ttc") || strings.HasSuffix(lower, ".otc") {
		srcs, err := text.NewGoTextFaceSourcesFromCollection(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if len(srcs) == 0 {
			return nil, fmt.Errorf("minigui: empty font collection: %s", path)
		}
		return srcs[0], nil
	}
	return text.NewGoTextFaceSource(bytes.NewReader(data))
}
