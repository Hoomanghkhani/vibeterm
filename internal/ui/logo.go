package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"vibeterm/internal/terminal"
)

// VibeTermLogo returns a sleek code-generated brand logo with neon cyan prompt '>_' and geometric purple 'V'
func VibeTermLogo() fyne.CanvasObject {
	promptSymbol := canvas.NewText(">_", terminal.ColorNeonCyan)
	promptSymbol.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	promptSymbol.TextSize = 16

	brandLetter := canvas.NewText("V", terminal.ColorBrandPurple)
	brandLetter.TextStyle = fyne.TextStyle{Bold: true}
	brandLetter.TextSize = 17

	brandName := canvas.NewText("IBETERM", terminal.ColorForeground)
	brandName.TextStyle = fyne.TextStyle{Bold: true}
	brandName.TextSize = 14

	logoRow := container.NewHBox(promptSymbol, brandLetter, brandName)
	return logoRow
}

// VibeTermBrandHeader returns the logo with the sleek italicized muted tagline
func VibeTermBrandHeader() fyne.CanvasObject {
	logo := VibeTermLogo()

	tagline := canvas.NewText("The Infrastructure Terminal with Good Vibes.", terminal.ColorDimmedText)
	tagline.TextSize = 10
	tagline.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		logo,
		tagline,
	)
}
