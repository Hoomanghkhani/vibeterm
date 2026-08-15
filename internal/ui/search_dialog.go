package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/models"
	"vibeterm/internal/terminal"
)

// ShowSearchDialog opens a dedicated quick search modal for hosts and snippets
func ShowSearchDialog(parent fyne.Window, hosts []models.Host, onSelectHost func(models.Host)) {
	w := fyne.CurrentApp().NewWindow("🔍 Quick Search & Connect (Ctrl+P)")
	w.Resize(fyne.NewSize(560, 420))

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Type host name, IP, tag, or environment...")

	filtered := hosts

	var list *widget.List
	list = widget.NewList(
		func() int {
			return len(filtered)
		},
		func() fyne.CanvasObject {
			dot := canvas.NewCircle(terminal.ColorGreen)
			dot.Resize(fyne.NewSize(6, 6))

			nameLabel := canvas.NewText("Host Name", terminal.ColorForeground)
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}
			nameLabel.TextSize = 13

			subLabel := canvas.NewText("user@host", terminal.ColorDimmedText)
			subLabel.TextSize = 10

			envLabel := canvas.NewText("PROD", terminal.ColorYellow)
			envLabel.TextSize = 10

			leftBox := container.NewHBox(container.NewPadded(dot), container.NewVBox(nameLabel, subLabel))
			return container.NewBorder(nil, nil, leftBox, container.NewPadded(envLabel))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(filtered) {
				return
			}
			h := filtered[i]
			border := o.(*fyne.Container)
			leftBox := border.Objects[0].(*fyne.Container)
			envLabel := border.Objects[1].(*fyne.Container).Objects[0].(*canvas.Text)

			dotContainer := leftBox.Objects[0].(*fyne.Container)
			dot := dotContainer.Objects[0].(*canvas.Circle)
			vbox := leftBox.Objects[1].(*fyne.Container)
			nameLabel := vbox.Objects[0].(*canvas.Text)
			subLabel := vbox.Objects[1].(*canvas.Text)

			nameLabel.Text = h.Name
			if h.Protocol == models.ProtocolLocal {
				subLabel.Text = "Local Shell"
				dot.FillColor = terminal.ColorNeonCyan
			} else {
				subLabel.Text = fmt.Sprintf("%s@%s:%d", h.Username, h.Hostname, h.Port)
				switch h.Health {
				case models.HealthOnline:
					dot.FillColor = terminal.ColorGreen
				case models.HealthDegraded:
					dot.FillColor = terminal.ColorYellow
				case models.HealthOffline:
					dot.FillColor = terminal.ColorRed
				default:
					dot.FillColor = terminal.ColorDimmedText
				}
			}

			envLabel.Text = strings.ToUpper(h.Environment)
			nameLabel.Refresh()
			subLabel.Refresh()
			envLabel.Refresh()
			dot.Refresh()
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(filtered) && onSelectHost != nil {
			onSelectHost(filtered[id])
			w.Close()
		}
	}

	searchEntry.OnChanged = func(q string) {
		q = strings.ToLower(strings.TrimSpace(q))
		if q == "" {
			filtered = hosts
		} else {
			var matched []models.Host
			for _, h := range hosts {
				if strings.Contains(strings.ToLower(h.Name), q) ||
					strings.Contains(strings.ToLower(h.Hostname), q) ||
					strings.Contains(strings.ToLower(h.Environment), q) {
					matched = append(matched, h)
					continue
				}
				for _, tag := range h.Tags {
					if strings.Contains(strings.ToLower(tag), q) {
						matched = append(matched, h)
						break
					}
				}
			}
			filtered = matched
		}
		list.Refresh()
	}

	w.SetContent(container.NewBorder(
		container.NewPadded(searchEntry),
		nil,
		nil,
		nil,
		list,
	))

	w.Canvas().Focus(searchEntry)
	w.Show()
}
