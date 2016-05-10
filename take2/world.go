package main

import "github.com/gdamore/tcell"

type World struct {
	Config  *Config
	Display *Display
}

func NewWorld() *World {
	return &World{}
}

func (w *World) Init() error {
	return nil
}

func (w *World) Run() {
	quit := make(chan struct{})

	go func() {
		for {
			/*
				w.Display.End()
				spew.Dump(w)
				if true {
					return
				}
			*/
			w.Display.Render()

			// Now wait for and handle user event
			ev := w.Display.Screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyCtrlQ:
					close(quit)
					return
				default:
					if ev.Key() == tcell.KeyRune {
						w.Display.CurrentBuffer().Insert(string(ev.Rune()))
					}
				}
			case *tcell.EventResize:
				w.Display.FullRender()
			}
		}
	}()

	<-quit
	w.Display.End()
}
