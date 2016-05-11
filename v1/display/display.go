package display

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/kiasaki/editing/config"
	"github.com/kiasaki/editing/text"
	"github.com/mattn/go-runewidth"
)

type Display struct {
	screen         tcell.Screen
	config         *config.Config
	windowTree     *Window
	currentWindow  *Window
	awayFromWindow bool
	world          *text.World
	width          int
	height         int
}

func DisplayNew(c *config.Config, w *text.World) *Display {
	return &Display{config: c, world: w, awayFromWindow: false, width: 80, height: 24}
}

func (d *Display) Init() (err error) {
	d.screen, err = tcell.NewScreen()
	if err != nil {
		return
	}

	encoding.Register()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	err = d.screen.Init()
	if err != nil {
		return
	}

	// Set initial window tree to one window showing current buffer
	d.windowTree = WindowNewNode(d.world.CurrentBuffer())
	d.currentWindow = d.windowTree

	d.screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	d.screen.Clear()

	return nil
}

func (d *Display) End() error {
	d.screen.Fini()
	return nil
}

func (d *Display) Screen() tcell.Screen {
	return d.screen
}

func (d *Display) CurrentWindow() *Window {
	return d.currentWindow
}

func (d *Display) SetCurrentWindow(window *Window) {
	if window == nil || window.kind != WindowNode {
		panic("Display.SetCurrentWindow: Current window must be a node")
	}
	d.currentWindow = window
	d.world.SetCurrentBuffer(window.Buffer().Name())
}

// Ensure current buffer contents are still visible on screen
// if not scroll down/up to prefered % (or center)
func (d *Display) Frame(force bool) {

}

// Internal redisplay logic used by Redisplay() and Refresh()
// TODO: Implement progrsive rediplay and stop if called too fast
func (d *Display) redisplay() {
	d.width, d.height = d.screen.Size()
	d.screen.Clear()

	//d.screen.ShowCursor(x, y)
	//d.screen.SetContent(x, y, rune, []rune{}, style)

	// (height-1) -> leave one line for the command bar
	d.displayWindowTree(d.windowTree, 0, 0, d.width, d.height-1)
}

func (d *Display) displayWindowTree(windowTree *Window, x int, y int, width int, height int) {
	switch d.windowTree.kind {
	case WindowNode:
		d.displayWindow(d.windowTree, x, y, width, height)
	case WindowHorizontalSplit:
		halfWidth := int(math.Floor(float64(width) / 2.0))
		d.displayWindowTree(windowTree.left, x, y, halfWidth, height)
		d.displayWindowTree(windowTree.right, (x + halfWidth), y, (width - halfWidth), height)
	case WindowVerticalSplit:
		halfHeight := int(math.Floor(float64(width) / 2.0))
		d.displayWindowTree(windowTree.left, x, y, width, halfHeight)
		d.displayWindowTree(windowTree.right, x, (y + halfHeight), width, (height - halfHeight))
	}
}

func (d *Display) displayWindow(windowNode *Window, x int, y int, width int, height int) {
	buffer := windowNode.buffer
	statusBarStyle := StringToStyle(d.config.GetColor("statusbar"))

	for _, line := range buffer.Contents() {
		if utf8.RuneCountInString(line) > width {
			// TODO handle breaking/spilling over lines
			line = line[:width-1]
		}
		d.puts(tcell.StyleDefault, x, y, line)
	}

	// Cursor
	if !d.awayFromWindow && d.currentWindow == windowNode {
		// TODO x - 1 when in normal mode (assuming insert default now)
		d.screen.ShowCursor(x+buffer.Point().Char+1, y+buffer.Point().Line-1)
	}

	d.puts(statusBarStyle, x, y+height-1, pad(buffer.Name(), width, ' '))
}

// Executes incremental redisplay stopping if
// user is still typing, then picking up
func (d *Display) Redisplay() {
	d.redisplay()
	d.screen.Show()
}

// Forced full display
func (d *Display) Refresh() {
	d.redisplay()
	d.screen.Sync()
}

// Works as Redisplay but also centers point to prefered screen %
func (d *Display) Recenter() {
	d.Frame(true)
	d.Redisplay()
}

// SetPreferredPercentage(perc)
// GetPointRow() - on screeen
// GetPointCol() - taking in account line wrap + ui
// WindowCreate(w)
// WindowDestroy(w)
// WindowGrow(w, amount)
// GetWindowTopLine(w) - on screen
// GetWindowBottomLine(w)
// GetWindowTop(w) Position - buffer position at visible top left
// GetWindowBottom(w) Position - buffer position at visible bottom right

func (d *Display) puts(style tcell.Style, x, y int, str string) {
	s := d.screen
	i := 0
	var deferred []rune
	dwidth := 0
	for _, r := range str {
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
}

func StringToStyle(str string) tcell.Style {
	var fg string
	bg := "default"
	split := strings.Split(str, ",")
	if len(split) > 1 {
		fg, bg = split[0], split[1]
	} else {
		fg = split[0]
	}
	fg = strings.TrimSpace(fg)
	bg = strings.TrimSpace(bg)

	style := tcell.StyleDefault.Foreground(StringToColor(fg)).Background(StringToColor(bg))
	if strings.Contains(str, "bold") {
		style = style.Bold(true)
	}
	if strings.Contains(str, "reverse") {
		style = style.Reverse(true)
	}
	if strings.Contains(str, "underline") {
		style = style.Underline(true)
	}
	return style
}

func StringToColor(str string) tcell.Color {
	switch str {
	case "black":
		return tcell.ColorBlack
	case "red":
		return tcell.ColorMaroon
	case "green":
		return tcell.ColorGreen
	case "yellow":
		return tcell.ColorOlive
	case "blue":
		return tcell.ColorNavy
	case "magenta":
		return tcell.ColorPurple
	case "cyan":
		return tcell.ColorTeal
	case "white":
		return tcell.ColorSilver
	case "brightblack", "lightblack":
		return tcell.ColorGray
	case "brightred", "lightred":
		return tcell.ColorRed
	case "brightgreen", "lightgreen":
		return tcell.ColorLime
	case "brightyellow", "lightyellow":
		return tcell.ColorYellow
	case "brightblue", "lightblue":
		return tcell.ColorBlue
	case "brightmagenta", "lightmagenta":
		return tcell.ColorFuchsia
	case "brightcyan", "lightcyan":
		return tcell.ColorAqua
	case "brightwhite", "lightwhite":
		return tcell.ColorWhite
	case "default":
		return tcell.ColorDefault
	default:
		// Check if this is a 256 color
		if num, err := strconv.Atoi(str); err == nil && num < 256 && num >= 0 {
			return tcell.Color(num)
		}

		// Probably a truecolor hex value
		return tcell.GetColor(str)
	}
}

func pad(str string, length int, padding rune) string {
	for len(str) < length {
		str = str + string(padding)
	}
	return str
}
