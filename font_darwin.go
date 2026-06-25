package minigui

// systemFontCandidates lists macOS font files to try, CJK-capable first so
// Japanese and other non-Latin text renders, with a Latin font as a last resort.
// The Hiragino entries are real macOS filenames (Japanese), not identifiers.
func systemFontCandidates() []string {
	return []string{
		"/System/Library/Fonts/ヒラギノ角ゴシック W4.ttc", // Hiragino Kaku Gothic, full Japanese
		"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
		"/System/Library/Fonts/Hiragino Sans GB.ttc", // CJK fallback
		"/System/Library/Fonts/PingFang.ttc",         // modern macOS CJK
		"/Library/Fonts/Arial Unicode.ttf",           // broad Unicode coverage when present
		"/System/Library/Fonts/Helvetica.ttc",        // Latin last resort
	}
}
