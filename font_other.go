//go:build !darwin && !windows && !linux

package minigui

// systemFontCandidates is empty on platforms without a known font location, so
// SystemFace returns an error and callers fall back to the built-in debug font.
func systemFontCandidates() []string {
	return nil
}
