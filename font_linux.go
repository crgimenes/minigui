package minigui

// systemFontCandidates lists common Linux font paths to try, CJK-capable first.
// Linux ships no guaranteed CJK font; if none match, the caller falls back to the
// built-in debug font.
func systemFontCandidates() []string {
	return []string{
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", // Latin last resort
	}
}
