package minigui

// systemFontCandidates lists Windows font files to try, CJK-capable first so
// Japanese and other non-Latin text renders, with a Latin font as a last resort.
func systemFontCandidates() []string {
	return []string{
		`C:\Windows\Fonts\YuGothM.ttc`,  // Yu Gothic Medium (Japanese)
		`C:\Windows\Fonts\meiryo.ttc`,   // Meiryo (Japanese)
		`C:\Windows\Fonts\msgothic.ttc`, // MS Gothic (Japanese, widely present)
		`C:\Windows\Fonts\segoeui.ttf`,  // Latin last resort
	}
}
