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
				case tcell.KeyEnter:
					w.Display.CurrentBuffer().NewLineAndIndent()
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					w.Display.CurrentBuffer().Backspace()
				case tcell.KeyDelete:
					w.Display.CurrentBuffer().Delete(1)
				case tcell.KeyLeft:
					w.Display.CurrentBuffer().PointMove(-1)
				case tcell.KeyRight:
					w.Display.CurrentBuffer().PointMove(1)
				default:
					if ev.Key() == tcell.KeyRune {
						w.Display.CurrentBuffer().Insert(string(ev.Rune()))
					} else if ev.Key() == tcell.KeySpace {
						w.Display.CurrentBuffer().Insert(" ")
					} else {
						w.Display.CurrentBuffer().Insert(ev.Name())
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
