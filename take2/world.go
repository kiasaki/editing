package main

import "github.com/gdamore/tcell"

type World struct {
	Config  *Config
	Display *Display
	Buffers []*Buffer
}

func NewWorld() *World {
	return &World{Buffers: []*Buffer{}}
}

func (w *World) Init() error {
	var err error
	w.Config, err = NewConfig()
	if err != nil {
		return err
	}

	w.Display, err = NewDisplay(w.Config)
	if err != nil {
		return err
	}

	return err
}

func (w *World) Run() {
	quit := make(chan struct{})

	go func() {
		for {
			if len(w.Buffers) == 0 {
				scratchBuffer := NewBuffer("*scratch*", "")
				w.Buffers = append(w.Buffers, scratchBuffer)
				w.Display.WindowTree = NewWindowNode(scratchBuffer)
				w.Display.SetCurrentWindow(w.Display.WindowTree)
			}

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
