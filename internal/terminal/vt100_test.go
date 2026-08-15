package terminal

import (
	"strings"
	"testing"
)

func TestVT100ScreenBuffer(t *testing.T) {
	sb := NewScreenBuffer(80, 24)

	// Write simple text
	sb.Write([]byte("Hello, VibeTerm!\r\nSecond Line"))

	grid, curX, curY := sb.GetSnapshot()
	if curY != 1 {
		t.Errorf("expected cursor Y = 1, got %d", curY)
	}
	if curX != 11 {
		t.Errorf("expected cursor X = 11, got %d", curX)
	}

	// Verify first row content
	firstRow := string([]rune{
		grid[0][0].Char, grid[0][1].Char, grid[0][2].Char, grid[0][3].Char, grid[0][4].Char,
	})
	if firstRow != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", firstRow)
	}

	plainText := sb.GetPlainText()
	if len(plainText) == 0 {
		t.Errorf("expected non-empty plain text")
	}
}

func TestVT100ANSIColorParsing(t *testing.T) {
	sb := NewScreenBuffer(80, 24)

	// Write text with ANSI red color: \x1b[31mRedText\x1b[0m
	sb.Write([]byte("\x1b[31mRedText\x1b[0m"))

	grid, _, _ := sb.GetSnapshot()
	cell := grid[0][0]
	if cell.Char != 'R' {
		t.Errorf("expected cell char 'R', got '%c'", cell.Char)
	}

	expectedRed := ANSIColors[1]
	if cell.Fg != expectedRed {
		t.Errorf("expected FG %v, got %v", expectedRed, cell.Fg)
	}
}

func TestVT100CursorAndClear(t *testing.T) {
	sb := NewScreenBuffer(80, 24)

	// Write text and clear display: \x1b[2J
	sb.Write([]byte("Some text\x1b[2J"))

	grid, curX, curY := sb.GetSnapshot()
	if curX != 0 || curY != 0 {
		t.Errorf("expected cursor reset to 0,0 after clear, got (%d,%d)", curX, curY)
	}
	if grid[0][0].Char != ' ' {
		t.Errorf("expected empty cell after clear screen, got '%c'", grid[0][0].Char)
	}
}

func TestVT100Scrollback(t *testing.T) {
	sb := NewScreenBuffer(40, 5)

	// Write 10 lines to trigger scrolling
	for i := 0; i < 10; i++ {
		sb.Write([]byte("Line\r\n"))
	}

	if len(sb.Scrollback) == 0 {
		t.Errorf("expected non-empty scrollback buffer after exceeding row limit")
	}
}

func TestVT100OSCSequenceInterception(t *testing.T) {
	sb := NewScreenBuffer(80, 24)

	// 1. Single chunk OSC with BEL (e.g. shell integration OSC 3008)
	sb.Write([]byte("\x1b]3008;start=8beefdead\x07Prompt> "))
	text := sb.GetPlainText()
	if strings.Contains(text, "3008") || strings.Contains(text, "8beefdead") {
		t.Errorf("OSC metadata leaked to screen buffer: %q", text)
	}
	if !strings.HasPrefix(text, "Prompt>") {
		t.Errorf("expected text 'Prompt>', got %q", text)
	}

	// 2. OSC with String Terminator (ST = \x1b\) (e.g. OSC 133 prompt markers)
	sb.Write([]byte("\x1b]133;A\x1b\\echo hi\r\n"))
	text2 := sb.GetPlainText()
	if strings.Contains(text2, "133") || strings.Contains(text2, ";A") {
		t.Errorf("OSC 133 metadata leaked to screen buffer: %q", text2)
	}
	if !strings.Contains(text2, "echo hi") {
		t.Errorf("expected 'echo hi' in text, got %q", text2)
	}

	// 3. Multi-chunk split OSC sequence
	sb2 := NewScreenBuffer(80, 24)
	sb2.Write([]byte("\x1b]3008;start="))
	sb2.Write([]byte("partial_hash_12345\x07CleanOutput"))
	text3 := sb2.GetPlainText()
	if strings.Contains(text3, "3008") || strings.Contains(text3, "partial_hash") {
		t.Errorf("Split chunk OSC metadata leaked: %q", text3)
	}
	if text3 != "CleanOutput" {
		t.Errorf("expected 'CleanOutput', got %q", text3)
	}

	// 4. Split ESC prefix across chunks
	sb3 := NewScreenBuffer(80, 24)
	sb3.Write([]byte("Pre\x1b"))
	sb3.Write([]byte("]0;window_title\x07Post"))
	text4 := sb3.GetPlainText()
	if strings.Contains(text4, "window_title") {
		t.Errorf("Split ESC OSC title leaked: %q", text4)
	}
	if text4 != "PrePost" {
		t.Errorf("expected 'PrePost', got %q", text4)
	}
}
