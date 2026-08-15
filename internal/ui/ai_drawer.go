package ui

import (
	"context"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/ai"
	"vibeterm/internal/terminal"
)

// AIDrawerView provides the native AI Copilot interaction drawer
type AIDrawerView struct {
	container      *fyne.Container
	promptInput    *widget.Entry
	outputArea     *widget.Entry
	providerSelect *widget.Select
	statusLabel    *widget.Label
	copilotService *ai.CopilotService
	OnInsertCmd    func(cmd string)
	OnClose        func()
}

// NewAIDrawerView creates a new AI Copilot drawer
func NewAIDrawerView(copilotService *ai.CopilotService, onInsert func(string), onClose func()) *AIDrawerView {
	drawer := &AIDrawerView{
		copilotService: copilotService,
		OnInsertCmd:    onInsert,
		OnClose:        onClose,
	}

	headerTitle := canvas.NewText("🤖 AI Copilot (Warp-Grade CLI)", terminal.ColorActiveBorder)
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerTitle.TextSize = 14

	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		if drawer.OnClose != nil {
			drawer.OnClose()
		}
	})

	headerRow := container.NewBorder(nil, nil, headerTitle, closeBtn)

	drawer.providerSelect = widget.NewSelect([]string{"ollama", "openai", "anthropic", "gemini"}, nil)
	drawer.providerSelect.SetSelected("ollama")

	drawer.promptInput = widget.NewMultiLineEntry()
	drawer.promptInput.SetPlaceHolder("Ask AI: e.g. 'Find all docker containers using > 1GB RAM' or 'Explain error'")
	drawer.promptInput.Wrapping = fyne.TextWrapWord
	drawer.promptInput.SetMinRowsVisible(3)

	drawer.outputArea = widget.NewMultiLineEntry()
	drawer.outputArea.SetPlaceHolder("AI generated commands and explanations will stream here...")
	drawer.outputArea.Wrapping = fyne.TextWrapWord
	drawer.outputArea.Disable()
	drawer.outputArea.SetMinRowsVisible(5)

	drawer.statusLabel = widget.NewLabel("Ready")
	drawer.statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	generateBtn := widget.NewButtonWithIcon("Generate", theme.MediaPlayIcon(), drawer.onGenerate)
	generateBtn.Importance = widget.HighImportance

	insertBtn := widget.NewButtonWithIcon("Insert to Terminal", theme.ContentPasteIcon(), func() {
		text := strings.TrimSpace(drawer.outputArea.Text)
		if text != "" {
			lines := strings.Split(text, "\n")
			firstLine := lines[0]
			if strings.HasPrefix(firstLine, "```") && len(lines) > 1 {
				firstLine = lines[1]
			}
			if drawer.OnInsertCmd != nil {
				drawer.OnInsertCmd(firstLine)
			}
		}
	})

	btnRow := container.NewHBox(drawer.providerSelect, generateBtn, insertBtn, drawer.statusLabel)

	drawer.container = container.NewVBox(
		headerRow,
		drawer.promptInput,
		btnRow,
		drawer.outputArea,
		widget.NewSeparator(),
	)

	return drawer
}

func (d *AIDrawerView) onGenerate() {
	prompt := strings.TrimSpace(d.promptInput.Text)
	if prompt == "" {
		return
	}

	d.statusLabel.SetText("Streaming from " + d.providerSelect.Selected + "...")
	d.outputArea.SetText("")

	go func() {
		ctx := context.Background()
		var fullOutput strings.Builder

		err := d.copilotService.StreamCompletion(ctx, ai.PromptRequest{
			Prompt:   prompt,
			Provider: d.providerSelect.Selected,
		}, func(chunk string) {
			fullOutput.WriteString(chunk)
			d.outputArea.SetText(fullOutput.String())
		})

		if err != nil {
			d.statusLabel.SetText("Error: " + err.Error())
			d.outputArea.SetText("Error generating completion:\n" + err.Error())
		} else {
			d.statusLabel.SetText("Done")
		}
	}()
}

// Container returns the layout container
func (d *AIDrawerView) Container() *fyne.Container {
	return d.container
}
