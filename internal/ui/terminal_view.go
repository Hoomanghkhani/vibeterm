package ui

import (
	"image/color"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"vibeterm/internal/terminal"
)

const (
	terminalPaddingLeft = float32(16.0)
	terminalPaddingTop  = float32(16.0)
	terminalLineHeight  = float32(18.0) // 1.3x multiplier with 13.5pt text
	terminalCharWidth   = float32(8.4)
)

// TerminalWidget is a pure native Go terminal canvas with ANSI parsing, keyboard input, and sub-millisecond redraws
type TerminalWidget struct {
	widget.BaseWidget
	buffer     *terminal.ScreenBuffer
	OnInput    func(data []byte)
	OnResize   func(cols, rows int)
	isFocused  bool
	charWidth  float32
	charHeight float32
	cursorTick bool
	mu         sync.Mutex
}

// NewTerminalWidget creates a new native terminal widget with Termius viewport ergonomics
func NewTerminalWidget(buffer *terminal.ScreenBuffer, onInput func([]byte), onResize func(int, int)) *TerminalWidget {
	tw := &TerminalWidget{
		buffer:     buffer,
		OnInput:    onInput,
		OnResize:   onResize,
		charWidth:  terminalCharWidth,
		charHeight: terminalLineHeight,
		cursorTick: true,
	}
	tw.ExtendBaseWidget(tw)

	// Smooth cursor blink ticker
	go func() {
		ticker := time.NewTicker(550 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			tw.cursorTick = !tw.cursorTick
			tw.Refresh()
		}
	}()

	return tw
}

// CreateRenderer builds the native GPU canvas renderer for the terminal
func (tw *TerminalWidget) CreateRenderer() fyne.WidgetRenderer {
	bgRect := canvas.NewRectangle(terminal.ColorEditorBg)
	cursorRect := canvas.NewRectangle(terminal.ColorCursor)

	r := &terminalRenderer{
		widget:     tw,
		background: bgRect,
		cursor:     cursorRect,
		textLines:  make([]*canvas.Text, 0),
		bgBlocks:   make([]*canvas.Rectangle, 0),
	}
	return r
}

type terminalRenderer struct {
	widget     *TerminalWidget
	background *canvas.Rectangle
	cursor     *canvas.Rectangle
	textLines  []*canvas.Text
	bgBlocks   []*canvas.Rectangle
	objects    []fyne.CanvasObject
	lastCols   int
	lastRows   int
}

func (r *terminalRenderer) Destroy() {}

func (r *terminalRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)

	// Compute effective width/height after 16px viewport padding
	availWidth := size.Width - (terminalPaddingLeft * 2)
	availHeight := size.Height - (terminalPaddingTop * 2)
	if availWidth < 100 {
		availWidth = 100
	}
	if availHeight < 50 {
		availHeight = 50
	}

	cw := r.widget.charWidth
	ch := r.widget.charHeight
	if cw <= 0 {
		cw = terminalCharWidth
	}
	if ch <= 0 {
		ch = terminalLineHeight
	}

	cols := int(availWidth / cw)
	rows := int(availHeight / ch)
	if cols < 20 {
		cols = 20
	}
	if rows < 5 {
		rows = 5
	}

	if (cols != r.lastCols || rows != r.lastRows) && (cols > 0 && rows > 0) {
		r.lastCols = cols
		r.lastRows = rows
		r.widget.buffer.Resize(cols, rows)
		if r.widget.OnResize != nil {
			go r.widget.OnResize(cols, rows)
		}
	}
}

func (r *terminalRenderer) MinSize() fyne.Size {
	return fyne.NewSize(380, 220)
}

func (r *terminalRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *terminalRenderer) Refresh() {
	grid, curX, curY := r.widget.buffer.GetSnapshot()
	rows := len(grid)
	if rows == 0 {
		return
	}
	cols := len(grid[0])

	cw := r.widget.charWidth
	ch := r.widget.charHeight

	var objs []fyne.CanvasObject
	objs = append(objs, r.background)

	// Draw text lines with individual ANSI SGR colors and 16px padding
	for rowIdx, row := range grid {
		var currentRun strings.Builder
		var currentFg color.NRGBA
		var currentBold bool
		var startCol int

		flushRun := func(endCol int) {
			if currentRun.Len() > 0 {
				txt := canvas.NewText(currentRun.String(), currentFg)
				txt.TextStyle = fyne.TextStyle{Monospace: true, Bold: currentBold}
				txt.TextSize = 13.5
				txt.Move(fyne.NewPos(terminalPaddingLeft+float32(startCol)*cw, float32(rowIdx)*ch+terminalPaddingTop))
				objs = append(objs, txt)
				currentRun.Reset()
			}
		}

		for c, cell := range row {
			chRune := cell.Char
			if chRune == 0 {
				chRune = ' '
			}
			fg := cell.Fg
			if fg.A == 0 {
				fg = terminal.ColorForeground
			}

			// Render non-default cell background
			if cell.Bg != terminal.ColorEditorBg && cell.Bg.A > 0 {
				bgBlock := canvas.NewRectangle(cell.Bg)
				bgBlock.Move(fyne.NewPos(terminalPaddingLeft+float32(c)*cw, float32(rowIdx)*ch+terminalPaddingTop))
				bgBlock.Resize(fyne.NewSize(cw, ch))
				objs = append(objs, bgBlock)
			}

			if currentRun.Len() == 0 {
				currentFg = fg
				currentBold = cell.Bold
				startCol = c
				currentRun.WriteRune(chRune)
			} else if fg == currentFg && cell.Bold == currentBold {
				currentRun.WriteRune(chRune)
			} else {
				flushRun(c)
				currentFg = fg
				currentBold = cell.Bold
				startCol = c
				currentRun.WriteRune(chRune)
			}
		}
		flushRun(len(row))
	}

	// Draw Cursor with neon glow
	if r.widget.cursorTick && curX < cols && curY < rows {
		r.cursor.FillColor = terminal.ColorCursor
		r.cursor.Move(fyne.NewPos(float32(curX)*cw+terminalPaddingLeft, float32(curY)*ch+terminalPaddingTop))
		r.cursor.Resize(fyne.NewSize(cw, ch))
		objs = append(objs, r.cursor)
	}

	r.objects = objs
	canvas.Refresh(r.widget)
}

// Focus handling
func (tw *TerminalWidget) FocusGained() {
	tw.isFocused = true
	tw.Refresh()
}

func (tw *TerminalWidget) FocusLost() {
	tw.isFocused = false
	tw.Refresh()
}

func (tw *TerminalWidget) Focused() bool {
	return tw.isFocused
}

func (tw *TerminalWidget) Tapped(ev *fyne.PointEvent) {
	if !tw.isFocused {
		if c := fyne.CurrentApp().Driver().CanvasForObject(tw); c != nil {
			c.Focus(tw)
		}
	}
}

// TypedRune captures standard character inputs
func (tw *TerminalWidget) TypedRune(r rune) {
	if tw.OnInput != nil {
		tw.OnInput([]byte(string(r)))
	}
}

// TypedKey captures special keyboard keys (Enter, Backspace, Arrows, Tab, Esc)
func (tw *TerminalWidget) TypedKey(ev *fyne.KeyEvent) {
	if tw.OnInput == nil {
		return
	}

	switch ev.Name {
	case fyne.KeyReturn, fyne.KeyEnter:
		tw.OnInput([]byte("\r"))
	case fyne.KeyBackspace:
		tw.OnInput([]byte{0x7F})
	case fyne.KeyTab:
		tw.OnInput([]byte("\t"))
	case fyne.KeyEscape:
		tw.OnInput([]byte{0x1b})
	case fyne.KeyUp:
		tw.OnInput([]byte("\x1b[A"))
	case fyne.KeyDown:
		tw.OnInput([]byte("\x1b[B"))
	case fyne.KeyRight:
		tw.OnInput([]byte("\x1b[C"))
	case fyne.KeyLeft:
		tw.OnInput([]byte("\x1b[D"))
	case fyne.KeyHome:
		tw.OnInput([]byte("\x1b[H"))
	case fyne.KeyEnd:
		tw.OnInput([]byte("\x1b[F"))
	case fyne.KeyPageUp:
		tw.OnInput([]byte("\x1b[5~"))
	case fyne.KeyPageDown:
		tw.OnInput([]byte("\x1b[6~"))
	case fyne.KeyDelete:
		tw.OnInput([]byte("\x1b[3~"))
	}
}

// TypedShortcut handles Ctrl+C, Ctrl+D, Ctrl+L, Ctrl+Z, Ctrl+V, etc.
func (tw *TerminalWidget) TypedShortcut(shortcut fyne.Shortcut) {
	if tw.OnInput == nil {
		return
	}

	if custom, ok := shortcut.(*desktop.CustomShortcut); ok {
		if custom.Modifier == fyne.KeyModifierControl {
			switch custom.KeyName {
			case fyne.KeyC:
				tw.OnInput([]byte{0x03}) // Ctrl+C (SIGINT)
			case fyne.KeyD:
				tw.OnInput([]byte{0x04}) // Ctrl+D (EOF)
			case fyne.KeyZ:
				tw.OnInput([]byte{0x1a}) // Ctrl+Z (SIGTSTP)
			case fyne.KeyL:
				tw.OnInput([]byte{0x0c}) // Ctrl+L (Clear)
			case fyne.KeyA:
				tw.OnInput([]byte{0x01}) // Ctrl+A (Beginning of line)
			case fyne.KeyE:
				tw.OnInput([]byte{0x05}) // Ctrl+E (End of line)
			case fyne.KeyU:
				tw.OnInput([]byte{0x15}) // Ctrl+U (Erase line)
			case fyne.KeyK:
				tw.OnInput([]byte{0x0b}) // Ctrl+K (Kill to end)
			case fyne.KeyW:
				tw.OnInput([]byte{0x17}) // Ctrl+W (Erase word)
			}
		}
	}
}

func (tw *TerminalWidget) Buffer() *terminal.ScreenBuffer {
	return tw.buffer
}
