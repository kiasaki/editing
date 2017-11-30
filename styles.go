package main

import "github.com/gdamore/tcell"

func style(name string) tcell.Style {
	// TODO make table based and configurable
	if name == "message.error" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorMaroon)
	}
	if name == "statusbar" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.Color(6))
	}
	if name == "statusbar.highlight" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.Color(5))
	}
	if name == "linenumber" {
		return tcell.StyleDefault.
			Foreground(tcell.Color(6))
	}
	if name == "search" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorOlive)
	}
	if name == "visual" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.Color(0))
	}
	if name == "special" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorPurple)
	}
	if name == "text.string" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorOlive)
	}
	if name == "text.number" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorNavy)
	}
	if name == "text.comment" {
		return tcell.StyleDefault.
			Foreground(10)
	}
	if name == "text.reserved" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorPurple)
	}
	if name == "text.special" {
		return tcell.StyleDefault.
			Foreground(tcell.Color(6))
	}
	if name == "cursor" {
		return tcell.StyleDefault.Reverse(true)
	}
	return tcell.StyleDefault.Foreground(tcell.Color(0))
}
