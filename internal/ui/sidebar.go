package ui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/models"
	"vibeterm/internal/terminal"
)

// SidebarView displays the ultra-clean Termius-style host tree with right-click context menu editing
type SidebarView struct {
	container      *fyne.Container
	hostList       *widget.List
	allHosts       []models.Host
	filteredHosts  []models.Host
	parentWindow   fyne.Window
	OnHostSelected func(host models.Host)
	OnAddHost      func()
	OnOpenScanner  func()
	OnOpenTunnels  func()
	OnEditHost     func(host models.Host)
	OnDeleteHost   func(hostID string)
	OnOpenIDE      func(host models.Host)
}

// HostRowWidget is an interactive host row supporting left-click connect and right-click context menu
type HostRowWidget struct {
	widget.BaseWidget
	host         models.Host
	sidebar      *SidebarView
	dot          *canvas.Circle
	nameLabel    *canvas.Text
	subLabel     *canvas.Text
	latencyLabel *canvas.Text
	bgRect       *canvas.Rectangle
	hovered      bool
}

var _ fyne.Tappable = (*HostRowWidget)(nil)
var _ fyne.SecondaryTappable = (*HostRowWidget)(nil)
var _ desktop.Hoverable = (*HostRowWidget)(nil)

func NewHostRowWidget(sidebar *SidebarView) *HostRowWidget {
	w := &HostRowWidget{
		sidebar: sidebar,
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *HostRowWidget) SetHost(h models.Host) {
	w.host = h
	w.Refresh()
}

func (w *HostRowWidget) Tapped(ev *fyne.PointEvent) {
	if w.sidebar.OnHostSelected != nil {
		w.sidebar.OnHostSelected(w.host)
	}
}

func (w *HostRowWidget) TappedSecondary(ev *fyne.PointEvent) {
	h := w.host

	menuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("⚡ Connect", func() {
			if w.sidebar.OnHostSelected != nil {
				w.sidebar.OnHostSelected(h)
			}
		}),
		fyne.NewMenuItem("📁 Open SFTP", func() {
			if w.sidebar.OnHostSelected != nil {
				w.sidebar.OnHostSelected(h)
			}
		}),
		fyne.NewMenuItem("💻 Open in VS Code / Cursor", func() {
			if w.sidebar.OnOpenIDE != nil {
				w.sidebar.OnOpenIDE(h)
			}
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("⚙️ Edit Host...", func() {
			if w.sidebar.OnEditHost != nil {
				w.sidebar.OnEditHost(h)
			}
		}),
		fyne.NewMenuItem("🗑️ Delete Host", func() {
			if w.sidebar.OnDeleteHost != nil {
				w.sidebar.OnDeleteHost(h.ID)
			}
		}),
	}

	menu := fyne.NewMenu(h.Name, menuItems...)
	if c := fyne.CurrentApp().Driver().CanvasForObject(w); c != nil {
		widget.ShowPopUpMenuAtPosition(menu, c, ev.AbsolutePosition)
	}
}

func (w *HostRowWidget) MouseIn(ev *desktop.MouseEvent) {
	w.hovered = true
	w.Refresh()
}

func (w *HostRowWidget) MouseOut() {
	w.hovered = false
	w.Refresh()
}

func (w *HostRowWidget) MouseMoved(ev *desktop.MouseEvent) {}

func (w *HostRowWidget) CreateRenderer() fyne.WidgetRenderer {
	w.bgRect = canvas.NewRectangle(color.Transparent)

	w.dot = canvas.NewCircle(terminal.ColorGreen)
	w.dot.Resize(fyne.NewSize(6, 6))

	w.nameLabel = canvas.NewText("", terminal.ColorForeground)
	w.nameLabel.TextStyle = fyne.TextStyle{Bold: true}
	w.nameLabel.TextSize = 13

	w.subLabel = canvas.NewText("", terminal.ColorDimmedText)
	w.subLabel.TextSize = 10

	w.latencyLabel = canvas.NewText("", terminal.ColorDimmedText)
	w.latencyLabel.TextStyle = fyne.TextStyle{Monospace: true}
	w.latencyLabel.TextSize = 11

	r := &hostRowRenderer{
		widget:       w,
		bgRect:       w.bgRect,
		dot:          w.dot,
		nameLabel:    w.nameLabel,
		subLabel:     w.subLabel,
		latencyLabel: w.latencyLabel,
	}
	return r
}

type hostRowRenderer struct {
	widget       *HostRowWidget
	bgRect       *canvas.Rectangle
	dot          *canvas.Circle
	nameLabel    *canvas.Text
	subLabel     *canvas.Text
	latencyLabel *canvas.Text
}

func (r *hostRowRenderer) Destroy() {}

func (r *hostRowRenderer) Layout(size fyne.Size) {
	r.bgRect.Resize(size)

	// Dot position
	r.dot.Move(fyne.NewPos(8, (size.Height-6)/2))
	r.dot.Resize(fyne.NewSize(6, 6))

	// Name & Subtitle
	r.nameLabel.Move(fyne.NewPos(22, 4))
	r.subLabel.Move(fyne.NewPos(22, 20))

	// Latency right aligned
	latWidth := float32(50)
	r.latencyLabel.Move(fyne.NewPos(size.Width-latWidth-8, (size.Height-14)/2))
	r.latencyLabel.Resize(fyne.NewSize(latWidth, 14))
}

func (r *hostRowRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 38)
}

func (r *hostRowRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bgRect, r.dot, r.nameLabel, r.subLabel, r.latencyLabel}
}

func (r *hostRowRenderer) Refresh() {
	h := r.widget.host

	if r.widget.hovered {
		r.bgRect.FillColor = color.NRGBA{R: 24, G: 32, B: 46, A: 180}
	} else {
		r.bgRect.FillColor = color.Transparent
	}

	r.nameLabel.Text = h.Name
	if h.Protocol == models.ProtocolLocal {
		r.subLabel.Text = "Local Shell"
		r.dot.FillColor = terminal.ColorNeonCyan
	} else {
		r.subLabel.Text = fmt.Sprintf("%s@%s", h.Username, h.Hostname)
		switch h.Health {
		case models.HealthOnline:
			r.dot.FillColor = terminal.ColorGreen
		case models.HealthDegraded:
			r.dot.FillColor = terminal.ColorYellow
		case models.HealthOffline:
			r.dot.FillColor = terminal.ColorRed
		default:
			r.dot.FillColor = terminal.ColorDimmedText
		}
	}

	if h.LatencyMs > 0 {
		r.latencyLabel.Text = fmt.Sprintf("%.1fms", h.LatencyMs)
	} else {
		r.latencyLabel.Text = ""
	}

	r.bgRect.Refresh()
	r.dot.Refresh()
	r.nameLabel.Refresh()
	r.subLabel.Refresh()
	r.latencyLabel.Refresh()
}

// NewSidebarView constructs the clean sidebar component without search bar clutter
func NewSidebarView(
	parentWin fyne.Window,
	hosts []models.Host,
	onSelect func(models.Host),
	onAdd func(),
	onScan func(),
	onTunnels func(),
	onEdit func(models.Host),
	onDelete func(string),
	onOpenIDE func(models.Host),
) *SidebarView {
	sb := &SidebarView{
		parentWindow:   parentWin,
		allHosts:       hosts,
		filteredHosts:  hosts,
		OnHostSelected: onSelect,
		OnAddHost:      onAdd,
		OnOpenScanner:  onScan,
		OnOpenTunnels:  onTunnels,
		OnEditHost:     onEdit,
		OnDeleteHost:   onDelete,
		OnOpenIDE:      onOpenIDE,
	}

	// Brand Header
	brandHeader := VibeTermBrandHeader()

	// Section Title: "HOSTS" and "+" New Host Action
	sectionTitle := canvas.NewText("HOSTS", terminal.ColorDimmedText)
	sectionTitle.TextStyle = fyne.TextStyle{Bold: true}
	sectionTitle.TextSize = 11

	addBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		if sb.OnAddHost != nil {
			sb.OnAddHost()
		}
	})

	headerRow := container.NewBorder(nil, nil, sectionTitle, addBtn)

	// Clean Host Tree List
	sb.hostList = widget.NewList(
		func() int {
			return len(sb.filteredHosts)
		},
		func() fyne.CanvasObject {
			return NewHostRowWidget(sb)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(sb.filteredHosts) {
				return
			}
			row := o.(*HostRowWidget)
			row.SetHost(sb.filteredHosts[i])
		},
	)

	topContainer := container.NewVBox(
		container.NewPadded(brandHeader),
		container.NewPadded(headerRow),
		widget.NewSeparator(),
	)

	sb.container = container.NewBorder(
		topContainer,
		nil,
		nil,
		nil,
		sb.hostList,
	)

	return sb
}

// UpdateHosts refreshes the host list
func (sb *SidebarView) UpdateHosts(hosts []models.Host) {
	sb.allHosts = hosts
	sb.filteredHosts = hosts
	sb.hostList.Refresh()
}

// FilterHosts updates the filtered view based on search query
func (sb *SidebarView) FilterHosts(q string) {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		sb.filteredHosts = sb.allHosts
	} else {
		var matched []models.Host
		for _, h := range sb.allHosts {
			if strings.Contains(strings.ToLower(h.Name), q) ||
				strings.Contains(strings.ToLower(h.Hostname), q) ||
				strings.Contains(strings.ToLower(h.Folder), q) ||
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
		sb.filteredHosts = matched
	}
	sb.hostList.Refresh()
}

// Container returns the layout container
func (sb *SidebarView) Container() *fyne.Container {
	return sb.container
}
