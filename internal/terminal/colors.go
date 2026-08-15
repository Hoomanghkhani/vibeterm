package terminal

import "image/color"

// Termius Deep High-Contrast Color Palette
var (
	ColorEditorBg       = color.NRGBA{R: 14, G: 16, B: 21, A: 255}    // #0E1015 (Deep Abyss Black)
	ColorSidebarBg      = color.NRGBA{R: 21, G: 24, B: 33, A: 255}    // #151821 (Dark Navy Gray)
	ColorActivityBarBg  = color.NRGBA{R: 21, G: 24, B: 33, A: 255}    // #151821
	ColorTabBarBg       = color.NRGBA{R: 21, G: 24, B: 33, A: 255}    // #151821
	ColorStatusBarBg    = color.NRGBA{R: 14, G: 16, B: 21, A: 255}    // #0E1015 (Seamless footer)
	ColorBorder         = color.NRGBA{R: 34, G: 38, B: 50, A: 255}    // #222632 (1px subtle divider)
	ColorNeonCyan       = color.NRGBA{R: 0, G: 245, B: 212, A: 255}   // #00F5D4 (Primary Accent)
	ColorBrandPurple    = color.NRGBA{R: 123, G: 97, B: 255, A: 255}  // #7B61FF (Brand Secondary)
	ColorNeonPurple     = ColorBrandPurple
	ColorAccentBlue     = color.NRGBA{R: 0, G: 245, B: 212, A: 255}   // #00F5D4
	ColorAccentHover    = color.NRGBA{R: 24, G: 32, B: 46, A: 255}    // #18202E
	ColorActiveBorder   = color.NRGBA{R: 0, G: 245, B: 212, A: 255}   // #00F5D4
	ColorForeground     = color.NRGBA{R: 248, G: 248, B: 242, A: 255} // #F8F8F2 (Crisp Off-White)
	ColorForegroundPure = color.NRGBA{R: 255, G: 255, B: 255, A: 255} // #FFFFFF
	ColorDimmedText     = color.NRGBA{R: 160, G: 174, B: 192, A: 255} // #A0AEC0 (Readable Light Gray)
	ColorGreen          = color.NRGBA{R: 80, G: 250, B: 123, A: 255}  // #50FA7B (Vibrant Green)
	ColorYellow         = color.NRGBA{R: 241, G: 250, B: 140, A: 255} // #F1FA8C (Vibrant Yellow)
	ColorRed            = color.NRGBA{R: 255, G: 85, B: 85, A: 255}   // #FF5555 (Vibrant Red)
	ColorCursor         = color.NRGBA{R: 0, G: 245, B: 212, A: 220}   // Glowing Neon Cyan
	ColorObsidianBg     = ColorEditorBg                               // Backwards compatibility alias
	ColorHeaderBg       = ColorTabBarBg                               // Backwards compatibility alias
)

// ANSI standard 16 colors mapped to high-contrast vibrant terminal palette
var ANSIColors = [16]color.NRGBA{
	{R: 14, G: 16, B: 21, A: 255},     // 0: Black (#0E1015)
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
