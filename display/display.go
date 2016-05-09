package display

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/kiasaki/editing/config"
	"github.com/mattn/go-runewidth"
)

type Display struct {
	screen     tcell.Screen
	config     *config.Config
	windowTree struct{}
	width      int
	height     int
}

func DisplayNew(c *config.Config) *Display {
	return &Display{config: c, width: 80, height: 24}
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

	y := d.height - 1
	for x := 0; x < d.width; x++ {
		statusBarStyle := StringToStyle(d.config.GetColor("statusbar"))
		d.puts(statusBarStyle, x, y, " ")
	}
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
		if num, err := strconv.Atoi(str); err == nil {
			return GetColor256(num)
		}
		// Probably a truecolor hex value
		return tcell.GetColor(str)
	}
}

// GetColor256 returns the tcell color for a number between 0 and 255
func GetColor256(color int) tcell.Color {
	colors := []tcell.Color{tcell.ColorBlack, tcell.ColorMaroon, tcell.ColorGreen,
		tcell.ColorOlive, tcell.ColorNavy, tcell.ColorPurple,
		tcell.ColorTeal, tcell.ColorSilver, tcell.ColorGray,
		tcell.ColorRed, tcell.ColorLime, tcell.ColorYellow,
		tcell.ColorBlue, tcell.ColorFuchsia, tcell.ColorAqua,
		tcell.ColorWhite, tcell.Color16, tcell.Color17, tcell.Color18, tcell.Color19, tcell.Color20,
		tcell.Color21, tcell.Color22, tcell.Color23, tcell.Color24, tcell.Color25, tcell.Color26, tcell.Color27, tcell.Color28,
		tcell.Color29, tcell.Color30, tcell.Color31, tcell.Color32, tcell.Color33, tcell.Color34, tcell.Color35, tcell.Color36,
		tcell.Color37, tcell.Color38, tcell.Color39, tcell.Color40, tcell.Color41, tcell.Color42, tcell.Color43, tcell.Color44,
		tcell.Color45, tcell.Color46, tcell.Color47, tcell.Color48, tcell.Color49, tcell.Color50, tcell.Color51, tcell.Color52,
		tcell.Color53, tcell.Color54, tcell.Color55, tcell.Color56, tcell.Color57, tcell.Color58, tcell.Color59, tcell.Color60,
		tcell.Color61, tcell.Color62, tcell.Color63, tcell.Color64, tcell.Color65, tcell.Color66, tcell.Color67, tcell.Color68,
		tcell.Color69, tcell.Color70, tcell.Color71, tcell.Color72, tcell.Color73, tcell.Color74, tcell.Color75, tcell.Color76,
		tcell.Color77, tcell.Color78, tcell.Color79, tcell.Color80, tcell.Color81, tcell.Color82, tcell.Color83, tcell.Color84,
		tcell.Color85, tcell.Color86, tcell.Color87, tcell.Color88, tcell.Color89, tcell.Color90, tcell.Color91, tcell.Color92,
		tcell.Color93, tcell.Color94, tcell.Color95, tcell.Color96, tcell.Color97, tcell.Color98, tcell.Color99, tcell.Color100,
		tcell.Color101, tcell.Color102, tcell.Color103, tcell.Color104, tcell.Color105, tcell.Color106, tcell.Color107, tcell.Color108,
		tcell.Color109, tcell.Color110, tcell.Color111, tcell.Color112, tcell.Color113, tcell.Color114, tcell.Color115, tcell.Color116,
		tcell.Color117, tcell.Color118, tcell.Color119, tcell.Color120, tcell.Color121, tcell.Color122, tcell.Color123, tcell.Color124,
		tcell.Color125, tcell.Color126, tcell.Color127, tcell.Color128, tcell.Color129, tcell.Color130, tcell.Color131, tcell.Color132,
		tcell.Color133, tcell.Color134, tcell.Color135, tcell.Color136, tcell.Color137, tcell.Color138, tcell.Color139, tcell.Color140,
		tcell.Color141, tcell.Color142, tcell.Color143, tcell.Color144, tcell.Color145, tcell.Color146, tcell.Color147, tcell.Color148,
		tcell.Color149, tcell.Color150, tcell.Color151, tcell.Color152, tcell.Color153, tcell.Color154, tcell.Color155, tcell.Color156,
		tcell.Color157, tcell.Color158, tcell.Color159, tcell.Color160, tcell.Color161, tcell.Color162, tcell.Color163, tcell.Color164,
		tcell.Color165, tcell.Color166, tcell.Color167, tcell.Color168, tcell.Color169, tcell.Color170, tcell.Color171, tcell.Color172,
		tcell.Color173, tcell.Color174, tcell.Color175, tcell.Color176, tcell.Color177, tcell.Color178, tcell.Color179, tcell.Color180,
		tcell.Color181, tcell.Color182, tcell.Color183, tcell.Color184, tcell.Color185, tcell.Color186, tcell.Color187, tcell.Color188,
		tcell.Color189, tcell.Color190, tcell.Color191, tcell.Color192, tcell.Color193, tcell.Color194, tcell.Color195, tcell.Color196,
		tcell.Color197, tcell.Color198, tcell.Color199, tcell.Color200, tcell.Color201, tcell.Color202, tcell.Color203, tcell.Color204,
		tcell.Color205, tcell.Color206, tcell.Color207, tcell.Color208, tcell.Color209, tcell.Color210, tcell.Color211, tcell.Color212,
		tcell.Color213, tcell.Color214, tcell.Color215, tcell.Color216, tcell.Color217, tcell.Color218, tcell.Color219, tcell.Color220,
		tcell.Color221, tcell.Color222, tcell.Color223, tcell.Color224, tcell.Color225, tcell.Color226, tcell.Color227, tcell.Color228,
		tcell.Color229, tcell.Color230, tcell.Color231, tcell.Color232, tcell.Color233, tcell.Color234, tcell.Color235, tcell.Color236,
		tcell.Color237, tcell.Color238, tcell.Color239, tcell.Color240, tcell.Color241, tcell.Color242, tcell.Color243, tcell.Color244,
		tcell.Color245, tcell.Color246, tcell.Color247, tcell.Color248, tcell.Color249, tcell.Color250, tcell.Color251, tcell.Color252,
		tcell.Color253, tcell.Color254, tcell.Color255,
	}

	return colors[color]
}
