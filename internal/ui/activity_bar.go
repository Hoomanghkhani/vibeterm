package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/terminal"
)

// ActivityIconButton is a custom minimalist activity bar icon with 2px accent indicator and zero button chrome
type ActivityIconButton struct {
	widget.BaseWidget
	iconRes   fyne.Resource
	active    bool
	hovered   bool
	onTapped  func()
	indicator *canvas.Rectangle
	icon      *widget.Icon
	bgRect    *canvas.Rectangle
}

var _ fyne.Tappable = (*ActivityIconButton)(nil)
var _ desktop.Hoverable = (*ActivityIconButton)(nil)

// NewActivityIconButton creates a transparent, sleek activity bar icon
func NewActivityIconButton(res fyne.Resource, onTapped func()) *ActivityIconButton {
	btn := &ActivityIconButton{
		iconRes:  res,
		onTapped: onTapped,
	}
	btn.ExtendBaseWidget(btn)
	return btn
}

func (b *ActivityIconButton) SetActive(active bool) {
	b.active = active
	b.Refresh()
}

func (b *ActivityIconButton) Tapped(ev *fyne.PointEvent) {
	if b.onTapped != nil {
		b.onTapped()
	}
}

func (b *ActivityIconButton) MouseIn(ev *desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

func (b *ActivityIconButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

func (b *ActivityIconButton) MouseMoved(ev *desktop.MouseEvent) {}

func (b *ActivityIconButton) CreateRenderer() fyne.WidgetRenderer {
	b.bgRect = canvas.NewRectangle(color.Transparent)
	b.indicator = canvas.NewRectangle(color.Transparent)
	b.indicator.Resize(fyne.NewSize(2, 32))

	b.icon = widget.NewIcon(b.iconRes)

	r := &activityIconRenderer{
		btn:       b,
		bgRect:    b.bgRect,
		indicator: b.indicator,
		icon:      b.icon,
	}
	return r
}

type activityIconRenderer struct {
	btn       *ActivityIconButton
	bgRect    *canvas.Rectangle
	indicator *canvas.Rectangle
	icon      *widget.Icon
}

func (r *activityIconRenderer) Destroy() {}

func (r *activityIconRenderer) Layout(size fyne.Size) {
	r.bgRect.Resize(size)

	// Left 2px accent bar
	r.indicator.Resize(fyne.NewSize(2, size.Height-12))
	r.indicator.Move(fyne.NewPos(0, 6))

	// Centered icon
	iconSize := float32(18)
	r.icon.Resize(fyne.NewSize(iconSize, iconSize))
	r.icon.Move(fyne.NewPos((size.Width-iconSize)/2+1, (size.Height-iconSize)/2))
}

func (r *activityIconRenderer) MinSize() fyne.Size {
	return fyne.NewSize(42, 40)
}

func (r *activityIconRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bgRect, r.indicator, r.icon}
}

func (r *activityIconRenderer) Refresh() {
	if r.btn.active {
		r.indicator.FillColor = terminal.ColorNeonCyan
		r.bgRect.FillColor = color.NRGBA{R: 24, G: 32, B: 46, A: 160}
	} else if r.btn.hovered {
		r.indicator.FillColor = color.Transparent
		r.bgRect.FillColor = color.NRGBA{R: 24, G: 32, B: 46, A: 120}
	} else {
		r.indicator.FillColor = color.Transparent
		r.bgRect.FillColor = color.Transparent
	}
	r.indicator.Refresh()
	r.bgRect.Refresh()
	r.icon.Refresh()
}

// ActivityBar creates the ultra-sleek far-left icon strip
type ActivityBar struct {
	container   *fyne.Container
	explorerBtn *ActivityIconButton
	tunnelBtn   *ActivityIconButton
	scannerBtn  *ActivityIconButton
	gitopsBtn   *ActivityIconButton
	helpBtn     *ActivityIconButton
	aboutBtn    *ActivityIconButton
}

func NewActivityBar(
	onExplorer func(),
	onTunnels func(),
	onScanner func(),
	onGitOps func(),
	onHelp func(),
	onAbout func(),
) *ActivityBar {
	ab := &ActivityBar{}

	ab.explorerBtn = NewActivityIconButton(theme.StorageIcon(), onExplorer)
	ab.explorerBtn.SetActive(true)

	ab.tunnelBtn = NewActivityIconButton(theme.NavigateNextIcon(), onTunnels)
	ab.scannerBtn = NewActivityIconButton(theme.SearchIcon(), onScanner)
	ab.gitopsBtn = NewActivityIconButton(theme.ViewRefreshIcon(), onGitOps)

	topStack := container.NewVBox(
		ab.explorerBtn,
		ab.tunnelBtn,
		ab.scannerBtn,
		ab.gitopsBtn,
	)

	ab.helpBtn = NewActivityIconButton(theme.DocumentIcon(), onHelp)
	ab.aboutBtn = NewActivityIconButton(theme.InfoIcon(), onAbout)

	bottomStack := container.NewVBox(
		ab.helpBtn,
		ab.aboutBtn,
	)

	bg := canvas.NewRectangle(terminal.ColorActivityBarBg)
	content := container.NewBorder(topStack, bottomStack, nil, nil)
	ab.container = container.NewStack(bg, container.NewPadded(content))

	return ab
}

func (ab *ActivityBar) SetActiveIndex(idx int) {
	ab.explorerBtn.SetActive(idx == 0)
	ab.tunnelBtn.SetActive(idx == 1)
	ab.scannerBtn.SetActive(idx == 2)
	ab.gitopsBtn.SetActive(idx == 3)
}

func (ab *ActivityBar) Container() *fyne.Container {
	return ab.container
}
