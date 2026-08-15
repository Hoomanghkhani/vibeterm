package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/forwarding"
	"vibeterm/internal/models"
	"vibeterm/internal/terminal"
)

// TunnelDrawerView displays active SSH port forwardings and SOCKS5 tunnels with live metrics
type TunnelDrawerView struct {
	container   *fyne.Container
	tunnelList  *widget.List
	activeRules []models.PortForwardRule
	orch        *forwarding.ForwardingOrchestrator
	OnClose     func()
}

// NewTunnelDrawerView creates a new Port Forwarding manager view
func NewTunnelDrawerView(orch *forwarding.ForwardingOrchestrator, onClose func()) *TunnelDrawerView {
	td := &TunnelDrawerView{
		orch:    orch,
		OnClose: onClose,
	}

	headerTitle := canvas.NewText("🔀 Active Port Forwarding & SOCKS5 Tunnels", terminal.ColorNeonCyan)
	headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	headerTitle.TextSize = 14

	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		if td.OnClose != nil {
			td.OnClose()
		}
	})

	headerRow := container.NewBorder(nil, nil, headerTitle, closeBtn)

	td.tunnelList = widget.NewList(
		func() int {
			return len(td.activeRules)
		},
		func() fyne.CanvasObject {
			nameLabel := widget.NewLabelWithStyle("Tunnel Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			typeBadge := canvas.NewText("LOCAL", terminal.ColorNeonCyan)
			typeBadge.TextStyle = fyne.TextStyle{Bold: true}
			routeLabel := widget.NewLabel("127.0.0.1:8080 -> 10.0.0.5:80")
			routeLabel.TextStyle = fyne.TextStyle{Monospace: true}
			metricsLabel := widget.NewLabel("Rx: 0 B | Tx: 0 B | Conns: 0")
			metricsLabel.TextStyle = fyne.TextStyle{Monospace: true}

			stopBtn := widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), nil)
			stopBtn.Importance = widget.DangerImportance

			infoBox := container.NewVBox(
				container.NewHBox(nameLabel, typeBadge),
				routeLabel,
				metricsLabel,
			)

			return container.NewBorder(nil, nil, infoBox, stopBtn)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(td.activeRules) {
				return
			}
			rule := td.activeRules[i]
			border := o.(*fyne.Container)
			infoBox := border.Objects[0].(*fyne.Container)
			stopBtn := border.Objects[1].(*widget.Button)

			topRow := infoBox.Objects[0].(*fyne.Container)
			nameLabel := topRow.Objects[0].(*widget.Label)
			typeBadge := topRow.Objects[1].(*canvas.Text)
			routeLabel := infoBox.Objects[1].(*widget.Label)
			metricsLabel := infoBox.Objects[2].(*widget.Label)

			nameLabel.SetText(rule.Name)
			typeBadge.Text = string(rule.Type)

			if rule.Type == models.ForwardDynamic {
				routeLabel.SetText(fmt.Sprintf("SOCKS5 Proxy on %s:%d", rule.BindAddress, rule.BindPort))
			} else {
				routeLabel.SetText(fmt.Sprintf("%s:%d -> %s:%d", rule.BindAddress, rule.BindPort, rule.TargetAddress, rule.TargetPort))
			}

			metricsLabel.SetText(fmt.Sprintf("Rx: %s | Tx: %s | Active Conns: %d", formatBytes(rule.RxBytes), formatBytes(rule.TxBytes), rule.ActiveConns))

			stopBtn.OnTapped = func() {
				td.orch.StopTunnel(rule.ID)
				td.Refresh()
			}
		},
	)

	// Refresh ticker for live bandwidth counters
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			td.Refresh()
		}
	}()

	td.container = container.NewBorder(
		container.NewVBox(headerRow, widget.NewSeparator()),
		nil,
		nil,
		nil,
		td.tunnelList,
	)

	return td
}

func (td *TunnelDrawerView) Refresh() {
	td.activeRules = td.orch.GetActiveTunnels()
	td.tunnelList.Refresh()
}

func (td *TunnelDrawerView) Container() *fyne.Container {
	return td.container
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
