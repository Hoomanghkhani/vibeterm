package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/terminal"
)

type shortcutEntry struct {
	Key         string
	Action      string
	Category    string
	Description string
}

var globalShortcuts = []shortcutEntry{
	{Key: "Ctrl + Shift + P / F1", Action: "Command Palette", Category: "General", Description: "Quickly access commands and actions"},
	{Key: "Ctrl + P", Action: "Quick Open / Search", Category: "General", Description: "Fuzzy search and connect to saved hosts"},
	{Key: "Ctrl + Shift + T / Ctrl + T", Action: "New Local Tab", Category: "Terminal", Description: "Launch a new local native shell session"},
	{Key: "Ctrl + W", Action: "Close Tab", Category: "Terminal", Description: "Terminate and close the active session tab"},
	{Key: "Ctrl + Shift + E", Action: "Split Vertically", Category: "Terminal", Description: "Split current workspace into side-by-side terminals"},
	{Key: "Ctrl + Shift + O", Action: "Split Horizontally", Category: "Terminal", Description: "Split current workspace into stacked terminals"},
	{Key: "Ctrl + I / Ctrl + K", Action: "Toggle AI Copilot", Category: "AI Copilot", Description: "Toggle non-intrusive AI CLI generation drawer"},
	{Key: "Ctrl + Shift + F", Action: "Global Search", Category: "General", Description: "Search hosts, tags, CIDR subnets, and snippets"},
	{Key: "Ctrl + L", Action: "Clear Terminal", Category: "Terminal", Description: "Clear terminal buffer and reset screen position"},
	{Key: "Ctrl + C", Action: "Interrupt (SIGINT)", Category: "Terminal", Description: "Send interrupt signal to running foreground process"},
	{Key: "Ctrl + D", Action: "EOF / Logout", Category: "Terminal", Description: "Send End-of-File signal or logout of remote session"},
}

// ShowHelpDialog opens the comprehensive Help & Keybindings modal
func ShowHelpDialog(parent fyne.Window) {
	w := fyne.CurrentApp().NewWindow("📖 VibeTerm — Help & Documentation")
	w.Resize(fyne.NewSize(760, 560))

	// Search bar for shortcuts & guides
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("🔍 Filter shortcuts or guides (e.g. 'split', 'bastion', 'socks5')...")

	// Shortcuts Tab Content
	shortcutsContainer := container.NewVBox()

	renderShortcuts := func(filter string) {
		shortcutsContainer.Objects = nil
		filter = strings.ToLower(strings.TrimSpace(filter))

		for _, sc := range globalShortcuts {
			if filter != "" {
				matches := strings.Contains(strings.ToLower(sc.Key), filter) ||
					strings.Contains(strings.ToLower(sc.Action), filter) ||
					strings.Contains(strings.ToLower(sc.Category), filter) ||
					strings.Contains(strings.ToLower(sc.Description), filter)
				if !matches {
					continue
				}
			}

			keyText := canvas.NewText(sc.Key, terminal.ColorActiveBorder)
			keyText.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
			keyText.TextSize = 13

			actionLabel := widget.NewLabelWithStyle(sc.Action, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			descLabel := widget.NewLabel(sc.Description)
			descLabel.Wrapping = fyne.TextWrapWord

			catBadge := canvas.NewText("["+sc.Category+"]", terminal.ColorYellow)
			catBadge.TextSize = 10

			rowLeft := container.NewHBox(keyText, catBadge)
			rowContent := container.NewVBox(actionLabel, descLabel)
			card := container.NewVBox(
				container.NewBorder(nil, nil, rowLeft, nil),
				rowContent,
				widget.NewSeparator(),
			)
			shortcutsContainer.Add(card)
		}
		shortcutsContainer.Refresh()
	}

	renderShortcuts("")
	searchEntry.OnChanged = renderShortcuts

	shortcutsScroll := container.NewVScroll(shortcutsContainer)
	shortcutsTab := container.NewBorder(searchEntry, nil, nil, nil, shortcutsScroll)

	// Quick Start Guides Tab Content
	guideText := `⚡ VibeTerm Architecture & Quick Start Guide

1. 🔗 Multi-Hop Bastion (Jump Host) Configuration
   - Navigate to Hosts Explorer and click 'New Host' (or right-click an existing host and choose 'Edit').
   - Open the 'Jump Bastion' section to chain arbitrary intermediate bastion gateways.
   - VibeTerm dials sequential TCP tunnels natively with independent authentication per hop.

2. 🔀 Port Forwarding Engine (Local, Remote, Dynamic SOCKS5)
   - Local Forward (-L): Binds a local port (e.g. 127.0.0.1:8080) and tunnels traffic to a remote service.
   - Remote Forward (-R): Listens on the remote SSH server and proxies inbound connections back locally.
   - Dynamic SOCKS5 (-D): Launches a local RFC 1928 compliant SOCKS5 proxy routed through the encrypted SSH pipe.
   - Real-time Rx/Tx telemetry meters transfer rates in the Port Forwarding drawer.

3. 🤖 Embedded AI Copilot (Warp-Grade CLI Assistant)
   - Press Ctrl+I or Ctrl+K to toggle the non-intrusive AI drawer.
   - Select your provider: Local Ollama (e.g. llama3), OpenAI (gpt-4o), Anthropic Claude, or Google Gemini.
   - Generate shell commands with terminal context injection, then click 'Insert to Terminal' to paste.

4. 🔄 GitOps Host & Snippet Synchronization
   - Connect your team's GitOps repository (GitHub, GitLab, or self-hosted Gitea).
   - Zero-Leak Security: All passwords, private keys, and passphrases are automatically stripped or vault-encrypted before commit.

5. 🛠️ External IDE Remote Bridge
   - Click 'Open in IDE' on any active SSH tab to trigger one-click remote attach in VS Code or Cursor:
     code --remote ssh-remote+<user>@<host> <remote_path>
`
	guideLabel := widget.NewLabel(guideText)
	guideLabel.Wrapping = fyne.TextWrapWord
	guideScroll := container.NewVScroll(container.NewPadded(guideLabel))

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("⌨️ Keybindings", theme.DocumentIcon(), shortcutsTab),
		container.NewTabItemWithIcon("🚀 Quick Start Guide", theme.InfoIcon(), guideScroll),
	)

	closeBtn := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() {
		w.Close()
	})
	closeBtn.Importance = widget.HighImportance

	bottomBar := container.NewHBox(closeBtn)
	w.SetContent(container.NewBorder(nil, container.NewPadded(bottomBar), nil, nil, tabs))
	w.Show()
}
