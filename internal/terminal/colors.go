package terminal

import "image/color"

// Termius Deep High-Contrast Color Palette
var (
	ColorEditorBg       = color.NRGBA{R: 30, G: 30, B: 30, A: 255}  // #1E1E1E (Cursor Editor)
	ColorSidebarBg      = color.NRGBA{R: 24, G: 24, B: 24, A: 255}  // #181818 (Cursor Sidebar)
	ColorActivityBarBg  = color.NRGBA{R: 24, G: 24, B: 24, A: 255}  // #181818
	ColorTabBarBg       = color.NRGBA{R: 24, G: 24, B: 24, A: 255}  // #181818
	ColorStatusBarBg    = color.NRGBA{R: 24, G: 24, B: 24, A: 255}  // #181818
	ColorBorder         = color.NRGBA{R: 51, G: 51, B: 51, A: 255}  // #333333 (Subtle divider)
	ColorNeonCyan       = color.NRGBA{R: 0, G: 127, B: 212, A: 255} // #007FD4 (Cursor Blue)
	ColorBrandPurple    = color.NRGBA{R: 0, G: 127, B: 212, A: 255} // #007FD4
	ColorNeonPurple     = ColorBrandPurple
	ColorAccentBlue     = color.NRGBA{R: 0, G: 127, B: 212, A: 255}   // #007FD4
	ColorAccentHover    = color.NRGBA{R: 42, G: 45, B: 46, A: 255}    // #2A2D2E
	ColorActiveBorder   = color.NRGBA{R: 0, G: 127, B: 212, A: 255}   // #007FD4
	ColorForeground     = color.NRGBA{R: 204, G: 204, B: 204, A: 255} // #CCCCCC (Off-White)
	ColorForegroundPure = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // #FFFFFF
	ColorDimmedText     = color.NRGBA{R: 133, G: 133, B: 133, A: 255} // #858585 (Dimmed)
	ColorGreen          = color.NRGBA{R: 137, G: 209, B: 133, A: 255} // #89D185 (Soft Green)
	ColorYellow         = color.NRGBA{R: 204, G: 167, B: 0, A: 255}   // #CCA700 (Soft Yellow)
	ColorRed            = color.NRGBA{R: 244, G: 135, B: 113, A: 255} // #F48771 (Soft Red)
	ColorCursor         = color.NRGBA{R: 0, G: 127, B: 212, A: 220}   // #007FD4
	ColorObsidianBg     = ColorEditorBg                               // Backwards compatibility alias
	ColorHeaderBg       = ColorTabBarBg                               // Backwards compatibility alias
)

// ANSI standard 16 colors mapped to high-contrast vibrant terminal palette
var ANSIColors = [16]color.NRGBA{
	{R: 14, G: 16, B: 21, A: 255},    // 0: Black (#0E1015)
	{R: 255, G: 85, B: 85, A: 255},   // 1: Red (#FF5555)
	{R: 80, G: 250, B: 123, A: 255},  // 2: Green (#50FA7B)
	{R: 241, G: 250, B: 140, A: 255}, // 3: Yellow (#F1FA8C)
	{R: 97, G: 175, B: 239, A: 255},  // 4: Blue (#61AFEF)
	{R: 189, G: 147, B: 249, A: 255}, // 5: Magenta (#BD93F9)
	{R: 139, G: 233, B: 253, A: 255}, // 6: Cyan (#8BE9FD)
	{R: 248, G: 248, B: 242, A: 255}, // 7: White (#F8F8F2)
	{R: 160, G: 174, B: 192, A: 255}, // 8: Bright Black (Grey #A0AEC0)
	{R: 255, G: 110, B: 110, A: 255}, // 9: Bright Red
	{R: 105, G: 255, B: 148, A: 255}, // 10: Bright Green
	{R: 255, G: 255, B: 165, A: 255}, // 11: Bright Yellow
	{R: 125, G: 190, B: 255, A: 255}, // 12: Bright Blue
	{R: 255, G: 121, B: 198, A: 255}, // 13: Bright Magenta
	{R: 0, G: 245, B: 212, A: 255},   // 14: Bright Cyan
	{R: 255, G: 255, B: 255, A: 255}, // 15: Bright White
}
