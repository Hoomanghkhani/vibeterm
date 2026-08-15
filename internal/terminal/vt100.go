package terminal

import (
	"image/color"
	"strconv"
	"strings"
	"sync"
)

// Cell represents a single character position in the terminal grid
type Cell struct {
	Char rune
	Fg   color.NRGBA
	Bg   color.NRGBA
	Bold bool
}

// ScreenBuffer maintains the 2D matrix of terminal cells and cursor state
type ScreenBuffer struct {
	mu         sync.RWMutex
	Cols       int
	Rows       int
	CursorX    int
	CursorY    int
	CurrentFg  color.NRGBA
	CurrentBg  color.NRGBA
	Bold       bool
	Grid       [][]Cell
	Scrollback [][]Cell
	MaxHistory int

	// Escape & OSC sequence parser state across chunks
	inOSC      bool
	inDCS      bool
	pendingEsc bool
}

// NewScreenBuffer initializes an empty terminal grid
func NewScreenBuffer(cols, rows int) *ScreenBuffer {
	if cols <= 0 {
		cols = 100
	}
	if rows <= 0 {
		rows = 30
	}

	sb := &ScreenBuffer{
		Cols:        cols,
		Rows:        rows,
		CursorX:     0,
		CursorY:     0,
		CurrentFg:   ColorForeground,
		CurrentBg:   ColorEditorBg,
		MaxHistory:  2000,
		Scrollback:  make([][]Cell, 0),
	}
	sb.Grid = make([][]Cell, rows)
	for r := range sb.Grid {
		sb.Grid[r] = make([]Cell, cols)
		for c := range sb.Grid[r] {
			sb.Grid[r][c] = Cell{Char: ' ', Fg: ColorForeground, Bg: ColorEditorBg}
		}
	}
	return sb
}

// Write processes incoming raw byte streams from PTY (parsing ANSI escapes & control chars)
func (sb *ScreenBuffer) Write(p []byte) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	i := 0
	length := len(p)

	for i < length {
		// If we are currently inside an OSC sequence, consume until terminator
		if sb.inOSC {
			termIdx, termLen := findOSCTerminator(p, i)
			if termIdx >= 0 {
				sb.inOSC = false
				i = termIdx + termLen
				continue
			} else {
				// Whole remaining slice is part of OSC sequence
				return
			}
		}

		// If we are inside a DCS/APC/PM sequence, consume until ST/BEL
		if sb.inDCS {
			termIdx, termLen := findOSCTerminator(p, i)
			if termIdx >= 0 {
				sb.inDCS = false
				i = termIdx + termLen
				continue
			} else {
				return
			}
		}

		b := p[i]

		// Handle pending standalone ESC from previous chunk
		if sb.pendingEsc {
			sb.pendingEsc = false
			i = sb.dispatchEscape(p, i-1, b)
			continue
		}

		// Escape sequence prefix
		if b == 0x1b {
			if i+1 >= length {
				// Standalone ESC at end of chunk
				sb.pendingEsc = true
				return
			}

			next := p[i+1]
			if next == '[' {
				// CSI sequence: parse until terminating byte (0x40..0x7E)
				j := i + 2
				for j < length && (p[j] < 0x40 || p[j] > 0x7E) {
					j++
				}
				if j < length {
					cmd := p[j]
					params := string(p[i+2 : j])
					sb.handleCSI(params, cmd)
					i = j + 1
					continue
				} else {
					// Incomplete CSI sequence at buffer boundary; consume safely
					return
				}
			} else if next == ']' {
				// OSC sequence: enter inOSC mode and look for terminator
				sb.inOSC = true
				termIdx, termLen := findOSCTerminator(p, i+2)
				if termIdx >= 0 {
					sb.inOSC = false
					i = termIdx + termLen
					continue
				} else {
					return
				}
			} else if next == 'P' || next == '_' || next == '^' {
				// DCS, APC, PM sequences
				sb.inDCS = true
				termIdx, termLen := findOSCTerminator(p, i+2)
				if termIdx >= 0 {
					sb.inDCS = false
					i = termIdx + termLen
					continue
				} else {
					return
				}
			} else if next == '(' || next == ')' || next == '*' || next == '+' {
				// Charset designation (e.g. \x1b(B)
				if i+2 < length {
					i += 3
				} else {
					i = length
				}
				continue
			} else {
				// 2-byte escape sequence: ESC c (Reset), ESC 7, ESC 8, ESC =, ESC >, etc.
				i += 2
				continue
			}
		}

		switch b {
		case '\r':
			sb.CursorX = 0
		case '\n':
			sb.newlineLocked()
		case '\b':
			if sb.CursorX > 0 {
				sb.CursorX--
			}
		case '\t':
			tabStop := ((sb.CursorX / 8) + 1) * 8
			if tabStop >= sb.Cols {
				tabStop = sb.Cols - 1
			}
			for sb.CursorX < tabStop {
				sb.putCharLocked(' ')
			}
		case 0x07: // BEL: ignore
		case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x0e, 0x0f: // C0 controls: ignore
		default:
			if b >= 32 {
				sb.putCharLocked(rune(b))
			}
		}
		i++
	}
}

// findOSCTerminator locates BEL (0x07) or ST (\x1b\ or 0x9c) in p starting at offset start
func findOSCTerminator(p []byte, start int) (termIdx int, termLen int) {
	length := len(p)
	for k := start; k < length; k++ {
		if p[k] == 0x07 {
			return k, 1
		}
		if p[k] == 0x9c {
			return k, 1
		}
		if p[k] == 0x1b && k+1 < length && p[k+1] == '\\' {
			return k, 2
		}
	}
	return -1, 0
}

func (sb *ScreenBuffer) dispatchEscape(p []byte, escIdx int, next byte) int {
	if next == '[' {
		// CSI
		j := 1
		length := len(p)
		for j < length && (p[j] < 0x40 || p[j] > 0x7E) {
			j++
		}
		if j < length {
			cmd := p[j]
			params := string(p[1:j])
			sb.handleCSI(params, cmd)
			return j + 1
		}
		return length
	}
	if next == ']' {
		sb.inOSC = true
		termIdx, termLen := findOSCTerminator(p, 1)
		if termIdx >= 0 {
			sb.inOSC = false
			return termIdx + termLen
		}
		return len(p)
	}
	return 1
}

func (sb *ScreenBuffer) putCharLocked(r rune) {
	if sb.CursorX >= sb.Cols {
		sb.CursorX = 0
		sb.newlineLocked()
	}
	if sb.CursorY >= sb.Rows {
		sb.CursorY = sb.Rows - 1
	}

	sb.Grid[sb.CursorY][sb.CursorX] = Cell{
		Char: r,
		Fg:   sb.CurrentFg,
		Bg:   sb.CurrentBg,
		Bold: sb.Bold,
	}
	sb.CursorX++
}

func (sb *ScreenBuffer) newlineLocked() {
	if sb.CursorY < sb.Rows-1 {
		sb.CursorY++
	} else {
		// Scroll up: save top line to scrollback
		if len(sb.Scrollback) >= sb.MaxHistory {
			sb.Scrollback = sb.Scrollback[1:]
		}
		sb.Scrollback = append(sb.Scrollback, sb.Grid[0])

		// Shift grid up
		for r := 0; r < sb.Rows-1; r++ {
			sb.Grid[r] = sb.Grid[r+1]
		}
		// Clear bottom row
		newRow := make([]Cell, sb.Cols)
		for c := range newRow {
			newRow[c] = Cell{Char: ' ', Fg: sb.CurrentFg, Bg: ColorEditorBg}
		}
		sb.Grid[sb.Rows-1] = newRow
	}
}

func (sb *ScreenBuffer) handleCSI(params string, cmd byte) {
	parts := strings.Split(params, ";")
	var args []int
	for _, p := range parts {
		// Strip private parameter prefixes like '?' in '\x1b[?2004h'
		cleanP := strings.TrimPrefix(p, "?")
		if val, err := strconv.Atoi(cleanP); err == nil {
			args = append(args, val)
		} else {
			args = append(args, 0)
		}
	}
	if len(args) == 0 {
		args = stringsToInts(params)
	}

	switch cmd {
	case 'm': // SGR: Select Graphic Rendition (Colors & Styles)
		if len(args) == 0 {
			sb.resetAttributes()
			return
		}
		for idx := 0; idx < len(args); idx++ {
			a := args[idx]
			switch {
			case a == 0:
				sb.resetAttributes()
			case a == 1:
				sb.Bold = true
			case a >= 30 && a <= 37:
				sb.CurrentFg = ANSIColors[a-30]
			case a >= 90 && a <= 97:
				sb.CurrentFg = ANSIColors[a-90+8]
			case a == 38: // Extended FG color
				if idx+4 < len(args) && args[idx+1] == 2 { // Truecolor RGB
					sb.CurrentFg = color.NRGBA{R: uint8(args[idx+2]), G: uint8(args[idx+3]), B: uint8(args[idx+4]), A: 255}
					idx += 4
				} else if idx+2 < len(args) && args[idx+1] == 5 { // 256 color
					if args[idx+2] < 16 {
						sb.CurrentFg = ANSIColors[args[idx+2]]
					}
					idx += 2
				}
			case a == 39:
				sb.CurrentFg = ColorForeground
			case a >= 40 && a <= 47:
				sb.CurrentBg = ANSIColors[a-40]
			case a >= 100 && a <= 107:
				sb.CurrentBg = ANSIColors[a-100+8]
			case a == 49:
				sb.CurrentBg = ColorEditorBg
			}
		}

	case 'H', 'f': // Cursor Position
		row := 1
		col := 1
		if len(args) > 0 && args[0] > 0 {
			row = args[0]
		}
		if len(args) > 1 && args[1] > 0 {
			col = args[1]
		}
		sb.CursorY = clamp(row-1, 0, sb.Rows-1)
		sb.CursorX = clamp(col-1, 0, sb.Cols-1)

	case 'J': // Erase in Display
		mode := 0
		if len(args) > 0 {
			mode = args[0]
		}
		if mode == 2 || mode == 3 { // Clear entire screen
			for r := range sb.Grid {
				for c := range sb.Grid[r] {
					sb.Grid[r][c] = Cell{Char: ' ', Fg: ColorForeground, Bg: ColorEditorBg}
				}
			}
			sb.CursorX = 0
			sb.CursorY = 0
		}

	case 'K': // Erase in Line
		for c := sb.CursorX; c < sb.Cols; c++ {
			sb.Grid[sb.CursorY][c] = Cell{Char: ' ', Fg: ColorForeground, Bg: ColorEditorBg}
		}

	case 'A': // Cursor Up
		n := 1
		if len(args) > 0 && args[0] > 0 {
			n = args[0]
		}
		sb.CursorY = clamp(sb.CursorY-n, 0, sb.Rows-1)

	case 'B': // Cursor Down
		n := 1
		if len(args) > 0 && args[0] > 0 {
			n = args[0]
		}
		sb.CursorY = clamp(sb.CursorY+n, 0, sb.Rows-1)

	case 'C': // Cursor Forward
		n := 1
		if len(args) > 0 && args[0] > 0 {
			n = args[0]
		}
		sb.CursorX = clamp(sb.CursorX+n, 0, sb.Cols-1)

	case 'D': // Cursor Back
		n := 1
		if len(args) > 0 && args[0] > 0 {
			n = args[0]
		}
		sb.CursorX = clamp(sb.CursorX-n, 0, sb.Cols-1)
	}
}

func (sb *ScreenBuffer) resetAttributes() {
	sb.CurrentFg = ColorForeground
	sb.CurrentBg = ColorEditorBg
	sb.Bold = false
}

// Resize updates the buffer dimensions and preserves existing lines
func (sb *ScreenBuffer) Resize(newCols, newRows int) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if newCols <= 0 || newRows <= 0 {
		return
	}

	newGrid := make([][]Cell, newRows)
	for r := range newGrid {
		newGrid[r] = make([]Cell, newCols)
		for c := range newGrid[r] {
			if r < sb.Rows && c < sb.Cols {
				newGrid[r][c] = sb.Grid[r][c]
			} else {
				newGrid[r][c] = Cell{Char: ' ', Fg: ColorForeground, Bg: ColorEditorBg}
			}
		}
	}

	sb.Cols = newCols
	sb.Rows = newRows
	sb.Grid = newGrid
	sb.CursorX = clamp(sb.CursorX, 0, newCols-1)
	sb.CursorY = clamp(sb.CursorY, 0, newRows-1)
}

// GetSnapshot returns a clone of current grid for rendering
func (sb *ScreenBuffer) GetSnapshot() ([][]Cell, int, int) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	gridClone := make([][]Cell, len(sb.Grid))
	for r := range sb.Grid {
		gridClone[r] = make([]Cell, len(sb.Grid[r]))
		copy(gridClone[r], sb.Grid[r])
	}
	return gridClone, sb.CursorX, sb.CursorY
}

// GetPlainText returns the full screen text (e.g. for AI copilot context)
func (sb *ScreenBuffer) GetPlainText() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	var sbText strings.Builder
	for _, row := range sb.Grid {
		line := make([]rune, len(row))
		for c, cell := range row {
			if cell.Char == 0 {
				line[c] = ' '
			} else {
				line[c] = cell.Char
			}
		}
		sbText.WriteString(strings.TrimRight(string(line), " "))
		sbText.WriteRune('\n')
	}
	return strings.TrimRight(sbText.String(), "\n")
}

func stringsToInts(s string) []int {
	if s == "" {
		return nil
	}
	var res []int
	for _, part := range strings.Split(s, ";") {
		if val, err := strconv.Atoi(part); err == nil {
			res = append(res, val)
		}
	}
	return res
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
