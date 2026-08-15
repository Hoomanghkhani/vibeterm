package ui

import (
	"fmt"
	"net/url"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/config"
)

// ShowAboutDialog displays the About Us and System Information modal
func ShowAboutDialog(parent fyne.Window) {
	w := fyne.CurrentApp().NewWindow("ℹ️ About VibeTerm")
	w.Resize(fyne.NewSize(580, 480))

	headerContainer := container.NewVBox(
		VibeTermBrandHeader(),
		widget.NewSeparator(),
	)

	// Version & Build Metadata
	versionStr := config.Version
	if versionStr == "" {
		versionStr = "1.0.0"
	}
	commitStr := config.GitCommit
	if commitStr == "" {
		commitStr = "release"
	}
	buildTimeStr := config.BuildTime
	if buildTimeStr == "" {
		buildTimeStr = "2026-08-15T00:00:00Z"
	}

	sysInfoText := fmt.Sprintf(`Version:       v%s
Git Commit:    %s
Build Time:    %s
Runtime:       %s (%s/%s)
Graphics:      Pure Native GPU Canvas (OpenGL/Vulkan)
Architecture:  Zero-Web / No-Electron / No-WebView / No-DOM`,
		versionStr,
		commitStr,
		buildTimeStr,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)

	sysInfoEntry := widget.NewMultiLineEntry()
	sysInfoEntry.SetText(sysInfoText)
	sysInfoEntry.TextStyle = fyne.TextStyle{Monospace: true}
	sysInfoEntry.Disable()
	sysInfoEntry.SetMinRowsVisible(6)

	copyBtn := widget.NewButtonWithIcon("Copy System Info", theme.ContentCopyIcon(), func() {
		w.Clipboard().SetContent(sysInfoText)
	})

	// Tech Stack & Credits
	techStackText := `🛠️ Built with High-Performance Pure Go:
  • GUI & Canvas Engine: Fyne v2.6.0 (Hardware-Accelerated Native Canvas)
  • Cryptography & SSH: golang.org/x/crypto/ssh, AES-256-GCM Vault
  • GitOps Synchronization: go-git/go-git/v5
  • Terminal State Machine: Pure Go VT100 / ANSI SGR Grid Buffer
  • AI Copilot Streaming: Ollama, OpenAI, Anthropic Claude, Google Gemini`

	techStackLabel := widget.NewLabel(techStackText)
	techStackLabel.Wrapping = fyne.TextWrapWord

	// External Links
	openURL := func(rawURL string) {
		u, err := url.Parse(rawURL)
		if err == nil {
			_ = fyne.CurrentApp().OpenURL(u)
		}
	}

	githubBtn := widget.NewButtonWithIcon("GitHub Repository", theme.HomeIcon(), func() {
		openURL("https://github.com/hoomanist/vibeterm")
	})

	docsBtn := widget.NewButtonWithIcon("Documentation", theme.DocumentIcon(), func() {
		openURL("https://github.com/hoomanist/vibeterm#readme")
	})

	issuesBtn := widget.NewButtonWithIcon("Report Issue", theme.HelpIcon(), func() {
		openURL("https://github.com/hoomanist/vibeterm/issues")
	})

	linksRow := container.NewHBox(githubBtn, docsBtn, issuesBtn)

	// License
	licenseLabel := widget.NewLabel("License: MIT License • Designed for Enterprise Infrastructure Teams")
	licenseLabel.TextStyle = fyne.TextStyle{Italic: true}

	closeBtn := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		w.Close()
	})
	closeBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		headerContainer,
		widget.NewLabelWithStyle("System Information", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sysInfoEntry,
		container.NewHBox(copyBtn),
		widget.NewSeparator(),
		techStackLabel,
		widget.NewSeparator(),
		linksRow,
		licenseLabel,
	)

	scroll := container.NewVScroll(container.NewPadded(content))
	bottomBar := container.NewHBox(closeBtn)

	w.SetContent(container.NewBorder(nil, container.NewPadded(bottomBar), nil, nil, scroll))
	w.Show()
}
