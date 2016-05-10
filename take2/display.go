package main

import (
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/mattn/go-runewidth"
)

type Display struct {
	Width  int
	Height int
	Screen tcell.Screen
}

func NewDisplay() (*Display, error) {
	var err error
	display := &Display{}

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
		Background(tcell.ColorBlack))
	display.Screen.Clear()

	return display, nil
}

func (d *Display) End() {
	d.Screen.Fini()
}

func (d *Display) CurrentBuffer() *Buffer {
	return nil
}

func (d *Display) render() {
	d.Width, d.Height = d.Screen.Size()
	d.Screen.Clear()
	d.write(tcell.StyleDefault, 0, 0, "Hello World!")
}

func (d *Display) Render() {
	d.render()
	d.Screen.Show()
}

func (d *Display) FullRender() {
	d.render()
	d.Screen.Sync()
}

func (d *Display) write(style tcell.Style, x, y int, str string) {
	s := d.Screen
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
