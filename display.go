package main

import (
	"math"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/mattn/go-runewidth"
)

type Display struct {
	Width         int
	Height        int
	Screen        tcell.Screen
	Config        *Config
	World         *World
	WindowTree    *Window
	CurrentWindow *Window
	// Used when focus is on minibuffer
	AwayFromWindow bool
}

func NewDisplay(c *Config, w *World) (*Display, error) {
	var err error
	display := &Display{Config: c, World: w, AwayFromWindow: false}

	display.Screen, err = tcell.NewScreen()
	if err != nil {
		return display, err
	}

	encoding.Register()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	err = display.Screen.Init()
	if err != nil {
		return display, err
	}

	display.Screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorDefault))
	display.Screen.Clear()

	return display, nil
}

func (d *Display) End() {
	d.Screen.Fini()
}

func (b *Display) HandleEvent(w *World, key *Key) bool {
	return b.CurrentBuffer().HandleEvent(w, key)
}

func (d *Display) CurrentBuffer() *Buffer {
	return d.CurrentWindow.Buffer
}

func (d *Display) SetCurrentWindow(window *Window) {
	if window == nil || window.Kind != WindowNode {
		panic("Display.SetCurrentWindow: Current window must be a node")
	}
	d.CurrentWindow = window
}

func (d *Display) ReplaceCurrentWindow(cb func(w *Window) *Window) {
	d.WindowTree = d.replaceWindow(d.CurrentWindow, d.WindowTree, cb)
}

func (d *Display) replaceWindow(target *Window, w *Window, cb func(w *Window) *Window) *Window {
	switch w.Kind {
	case WindowNode:
		if w == target {
			return cb(w)
		} else {
			return w
		}
	case WindowHorizontalSplit:
		w.Top = d.replaceWindow(target, w.Top, cb)
		w.Bottom = d.replaceWindow(target, w.Bottom, cb)
		return w
	case WindowVerticalSplit:
		w.Left = d.replaceWindow(target, w.Left, cb)
		w.Right = d.replaceWindow(target, w.Right, cb)
		return w
	}
	panic("unreachable")
}

func (d *Display) Render() {
	d.render()
	d.Screen.Show()
}

func (d *Display) FullRender() {
	d.render()
	d.Screen.Sync()
}

func (d *Display) render() {
	d.Width, d.Height = d.Screen.Size()
	d.Screen.Clear()

	// (height-1) -> leave one line for the command bar
	d.displayWindowTree(d.WindowTree, 0, 0, d.Width, d.Height-1)
	d.displayMiniBuffer()
}

func (d *Display) displayMiniBuffer() {
	defaultStyle := StringToStyle(d.Config.GetColor("default"))
	keysTyped := d.World.lastKeys.String()
	if len(d.World.Command) > 0 {
		d.write(defaultStyle, 0, d.Height-1, d.World.Command)
		d.write(defaultStyle.Reverse(true), len(d.World.Command), d.Height-1, " ")
	}
	d.write(defaultStyle, d.Width-len(keysTyped)-1, d.Height-1, keysTyped)
}

func (d *Display) displayWindowTree(windowTree *Window, x int, y int, width int, height int) {
	switch windowTree.Kind {
	case WindowNode:
		d.displayWindow(windowTree, x, y, width, height)
	case WindowHorizontalSplit:
		halfHeight := int(math.Floor(float64(height) / 2.0))
		d.displayWindowTree(windowTree.Top, x, y, width, halfHeight)
		d.displayWindowTree(windowTree.Bottom, x, (y + halfHeight), width, (height - halfHeight))
	case WindowVerticalSplit:
		defaultStyle := StringToStyle(d.Config.GetColor("default"))
		halfWidth := int(math.Floor(float64(width) / 2.0))

		// add a separator line between both
		for i := y; i < y+height; i++ {
			d.write(defaultStyle, x+halfWidth, i, "|")
		}

		d.displayWindowTree(windowTree.Left, x, y, halfWidth, height)
		d.displayWindowTree(windowTree.Right, (x + halfWidth + 1), y, (width - halfWidth), height)
	}
}

func (d *Display) displayWindow(window *Window, x int, y int, width int, height int) {
	buffer := window.Buffer
	bufferCursorChar := buffer.Cursor.Char
	bufferCursorLine := buffer.Cursor.Line
	inCommandMode := buffer.Modes.IsEditingModeNamed("command")

	defaultStyle := StringToStyle(d.Config.GetColor("default"))
	lineNumberStyle := StringToStyle(d.Config.GetColor("line-number"))
	statusBarStyle := StringToStyle(d.Config.GetColor("statusbar"))

	leftFringePadding := 1
	leftFringeHasNumbers := false
	if showLineNumbers, ok := d.Config.GetSetting("numbers"); ok {
		if showLineNumbers.(bool) {
			leftFringePadding += len(strconv.Itoa(len(buffer.Lines))) + 1
			leftFringeHasNumbers = true
		}
	}

	// Only when focused, frame & show cursor
	if !d.AwayFromWindow && d.CurrentWindow == window {
		statusBarStyle = StringToStyle(d.Config.GetColor("statusbar-active"))
		window.Frame(height)
	}

	// TODO start at current top
	currentLine := 0
	currentChar := 0

	currentY := y
	for currentY < y+height-1 && currentLine < len(buffer.Lines) {
		currentX := leftFringePadding + x
		currentChar = 0

		// TODO Handle case where cursor is at line (chage bg style?)

		if leftFringeHasNumbers {
			fringeText := PadLeft(strconv.Itoa(currentY-y+1)+" ", leftFringePadding, ' ')
			d.write(lineNumberStyle, x, currentY, fringeText)
		}

		extraSpaceForCursorAtEOL := " "
		for _, char := range buffer.Lines[currentLine] + extraSpaceForCursorAtEOL {
			charStyle := defaultStyle

			if currentLine == bufferCursorLine-1 && currentChar == bufferCursorChar {
				if d.CurrentWindow == window && !inCommandMode {
					charStyle = charStyle.Reverse(true)
				}
			}

			if currentX < x+width {
				charCountAdded := d.write(charStyle, currentX, currentY, string(char))
				currentX += charCountAdded
			} else {
				// TODO line not done and we are at the end of window, handle reflows?
			}

			currentChar++
		}

		currentLine++
		currentY++
	}

	statusBarModesText := "(" + d.statusBarModesText(buffer) + ")"
	statusBarPosText := "(" + strconv.Itoa(bufferCursorLine) + ", " + strconv.Itoa(bufferCursorChar) + ")"
	statusBarText := " " + buffer.Name + " " + statusBarPosText + " " + statusBarModesText + " "
	goOverVerticalSplitSep := 1
	d.write(statusBarStyle, x, y+height-1, Pad(statusBarText, width+goOverVerticalSplitSep, ' '))
}

func (d *Display) statusBarModesText(buffer *Buffer) string {
	statusBarModeNames := []string{}
	// Editing mode
	if m := buffer.Modes.EditingMode(); m != nil {
		statusBarModeNames = append(statusBarModeNames, m.Name)
	}
	// Major mode
	if m := buffer.Modes.MajorMode(); m != nil {
		statusBarModeNames = append(statusBarModeNames, m.Name)
	}
	// Minor modes
	for _, m := range buffer.Modes.MinorModes() {
		statusBarModeNames = append(statusBarModeNames, m.Name)
	}
	return strings.Join(statusBarModeNames, " ")
}

func (d *Display) write(style tcell.Style, x, y int, str string) int {
	s := d.Screen
	i := 0
	var deferred []rune
	dwidth := 0
	for _, r := range str {
		// Handle tabs
		if r == '\t' {
			tabWidth := d.Config.Settings["tabwidth"].(int)

			// Print first tab char
			s.SetContent(x+i, y, '>', nil, style.Foreground(tcell.ColorAqua))
			i++

			// Add space till we reach tab column or tabWidth
			for j := 0; j < tabWidth-1 || i%tabWidth == 0; j++ {
				s.SetContent(x+i, y, ' ', nil, style)
				i++
			}

			deferred = nil
			continue
		}

		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}

	if len(deferred) != 0 {
		s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}

	// i is the real width of what we just outputed
	return i
}
