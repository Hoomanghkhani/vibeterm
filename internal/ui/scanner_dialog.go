package ui

import (
	"context"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/models"
	"vibeterm/internal/scanner"
)

// ShowScannerDialog opens the native subnet and port scanner modal
func ShowScannerDialog(parent fyne.Window, onImportHost func(models.Host)) {
	cidrEntry := widget.NewEntry()
	cidrEntry.SetText("192.168.1.0/24")

	portsEntry := widget.NewEntry()
	portsEntry.SetText("22, 3389, 5900, 80, 443")

	statusLabel := widget.NewLabel("Ready to scan")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	var discovered []models.DiscoveredDevice
	resultsList := widget.NewList(
		func() int {
			return len(discovered)
		},
		func() fyne.CanvasObject {
			ipLabel := widget.NewLabelWithStyle("192.168.1.50", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			svcLabel := widget.NewLabel(":22 (SSH-2.0-OpenSSH)")
			svcLabel.TextStyle = fyne.TextStyle{Monospace: true}
			rttLabel := widget.NewLabel("1.2ms")

			importBtn := widget.NewButtonWithIcon("Import", theme.ContentAddIcon(), nil)
			importBtn.Importance = widget.HighImportance

			infoBox := container.NewVBox(ipLabel, svcLabel)
			return container.NewBorder(nil, nil, infoBox, container.NewHBox(rttLabel, importBtn))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(discovered) {
				return
			}
			d := discovered[i]
			border := o.(*fyne.Container)
			infoBox := border.Objects[0].(*fyne.Container)
			rightBox := border.Objects[1].(*fyne.Container)

			ipLabel := infoBox.Objects[0].(*widget.Label)
			svcLabel := infoBox.Objects[1].(*widget.Label)
			rttLabel := rightBox.Objects[0].(*widget.Label)
			importBtn := rightBox.Objects[1].(*widget.Button)

			ipLabel.SetText(d.IP)
			svcLabel.SetText(strings.Join(d.Services, " | "))
			rttLabel.SetText(fmt.Sprintf("%.1fms", d.LatencyMs))

			importBtn.OnTapped = func() {
				newHost := models.Host{
					Name:        fmt.Sprintf("Discovered (%s)", d.IP),
					Hostname:    d.IP,
					Port:        22,
					Protocol:    models.ProtocolSSH,
					Username:    "root",
					AuthMethod:  models.AuthPassword,
					Environment: "dev",
					Folder:      "Discovered Nodes",
					Tags:        []string{"scanned", "imported"},
					Color:       "#00F0FF",
					Health:      models.HealthOnline,
					LatencyMs:   d.LatencyMs,
				}
				if onImportHost != nil {
					onImportHost(newHost)
				}
			}
		},
	)

	netScanner := scanner.NewNetworkScanner()
	scanBtn := widget.NewButtonWithIcon("Start Scan", theme.SearchIcon(), func() {
		cidr := strings.TrimSpace(cidrEntry.Text)
		if cidr == "" {
			return
		}
		statusLabel.SetText("Scanning " + cidr + "...")

		go func() {
			ctx := context.Background()
			res, err := netScanner.ScanCIDR(ctx, cidr, []int{22, 3389, 5900, 80, 443}, 128)
			if err != nil {
				statusLabel.SetText("Scan error: " + err.Error())
			} else {
				discovered = res
				statusLabel.SetText(fmt.Sprintf("Found %d active hosts", len(res)))
				resultsList.Refresh()
			}
		}()
	})
	scanBtn.Importance = widget.HighImportance

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("🌐 Network & Subnet Scanner", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewForm(
				widget.NewFormItem("Subnet CIDR", cidrEntry),
				widget.NewFormItem("Target Ports", portsEntry),
			),
			container.NewHBox(scanBtn, statusLabel),
			widget.NewSeparator(),
		),
		nil,
		nil,
		nil,
		resultsList,
	)

	d := dialog.NewCustom("Subnet Device Discovery", "Close", content, parent)
	d.Resize(fyne.NewSize(580, 450))
	d.Show()
}
