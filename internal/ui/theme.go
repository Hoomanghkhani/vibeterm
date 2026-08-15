package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"vibeterm/internal/terminal"
)

// TermiusDeepTheme implements fyne.Theme with crisp high-contrast off-white text and vibrant accent colors
type TermiusDeepTheme struct{}

var _ fyne.Theme = (*TermiusDeepTheme)(nil)

// NewVibeTermTheme creates a new custom Termius Deep theme
func NewVibeTermTheme() fyne.Theme {
	return &TermiusDeepTheme{}
}

func (t *TermiusDeepTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return terminal.ColorEditorBg // #0E1015 (Deep Abyss Black)
	case theme.ColorNameHeaderBackground, theme.ColorNameMenuBackground:
		return terminal.ColorTabBarBg // #151821 (Dark Navy Gray)
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 17, G: 19, B: 26, A: 255} // #11131A
	case theme.ColorNameButton:
		return color.NRGBA{R: 21, G: 24, B: 33, A: 255} // #151821
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 17, G: 19, B: 26, A: 255}
	case theme.ColorNamePrimary:
		return terminal.ColorNeonCyan // #00F5D4
	case theme.ColorNameFocus:
		return terminal.ColorNeonCyan // #00F5D4
	case theme.ColorNameHover:
		return terminal.ColorAccentHover // #18202E
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0, G: 245, B: 212, A: 50}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 160}
	case theme.ColorNameForeground:
		return terminal.ColorForeground // #F8F8F2 (Crisp Off-White)
	case theme.ColorNameDisabled:
		return terminal.ColorDimmedText // #A0AEC0 (Readable Light Gray)
	case theme.ColorNameSeparator:
		return terminal.ColorBorder // #222632
	case theme.ColorNameSuccess:
		return terminal.ColorGreen // #50FA7B
	case theme.ColorNameWarning:
		return terminal.ColorYellow // #F1FA8C
	case theme.ColorNameError:
		return terminal.ColorRed // #FF5555
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

func (t *TermiusDeepTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *TermiusDeepTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *TermiusDeepTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInlineIcon:
		return 16
	case theme.SizeNameScrollBar:
		return 6
	case theme.SizeNameText:
		return 13
	case theme.SizeNameHeadingText:
		return 16
	case theme.SizeNameSubHeadingText:
		return 13
	case theme.SizeNameCaptionText:
		return 11
	default:
		return theme.DefaultTheme().Size(name)
	}
}
