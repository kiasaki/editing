package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
)

func handlePanics() {
	err := recover()
	if err != nil {
		switch err := err.(type) {
		case error:
			Fatal(err)
		case string:
			Fatal(errors.New(err))
		default:
			Fatal(errors.New(fmt.Sprintf("Unknown panic type: %v", err)))
		}
	}
}

func Fatal(err error) {
	if world != nil && world.Display != nil {
		world.Display.End()
		fmt.Fprintf(os.Stderr, "%v\n", "ENDED CLEAN")
	}
	fmt.Fprintf(os.Stderr, "%v\n", err)
	fmt.Print(errors.Wrap(err, 2).ErrorStack())
	os.Exit(1)
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

func Pad(str string, length int, padding rune) string {
	for utf8.RuneCountInString(str) < length {
		str = str + string(padding)
	}
	return str
}

func PadLeft(str string, length int, padding rune) string {
	for utf8.RuneCountInString(str) < length {
		str = string(padding) + str
	}
	return str
}

func ReverseString(str string) string {
	newStr := ""
	for _, c := range str {
		newStr += string(c)
	}
	return newStr
}

type LocationComparison int

const (
	LocationBefore LocationComparison = iota
	LocationSame                      = iota
	LocationAfter                     = iota
)

func CompareLocations(p1 Location, p2 Location) LocationComparison {
	if p1 == p2 {
		return LocationSame
	} else if p1 < p2 {
		return LocationBefore
	} else {
		return LocationAfter
	}
}
