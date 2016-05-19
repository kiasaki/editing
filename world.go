package main

import "github.com/gdamore/tcell"

type World struct {
	quit    chan (struct{})
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

func (w *World) Quit() {
	if w.quit != nil {
		close(w.quit)
	}
}

func (w *World) Run() {
	w.quit = make(chan struct{})

	var ev tcell.Event

	go func() {
		lastKeys := NewKey("")

		for {
			if len(w.Buffers) == 0 {
				scratchBuffer := NewBuffer("*scratch*", "")
				w.Buffers = append(w.Buffers, scratchBuffer)
				w.Display.WindowTree = NewWindowNode(scratchBuffer)
				w.Display.SetCurrentWindow(w.Display.WindowTree)
			}

			w.Display.Render()
			w.Display.write(tcell.StyleDefault, 10, 10, lastKeys.String())
			switch ev := ev.(type) {
			case *tcell.EventKey:
				w.Display.write(tcell.StyleDefault, 10, 12, ev.Name())
			}
			w.Display.Screen.Show()

			// Now wait for and handle user event
			ev = w.Display.Screen.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlQ {
					// TODO remove safeguard quit once bindings
					// implentation works well
					w.Quit()
					return
				} else if ev.Key() == tcell.KeyCtrlG {
					// Cancel entered keys
					lastKeys = NewKey("")
				} else if ev.Key() == tcell.KeyEscape && lastKeys.Length() > 0 {
					// Cancel entered keys
					lastKeys = NewKey("")
				} else {
					// Add lastest key store to what was already types and
					// check if a binding will handle it.
					// If used up, reset types keys
					keyStroke := NewKeyStrokeFromKeyEvent(ev)
					lastKeys.AppendKeyStroke(keyStroke)

					// Loop on the set of keys last entered and see if it matches any
					// bound function for the current mode
					// "C-a b c M-x" might not be bound, neither is "b c M-x" but if
					// get to just the last key "M-x" is bound.
					for i := 0; i < len(lastKeys.keys); i++ {
						didUseKey := w.Display.HandleEvent(w, &Key{lastKeys.keys[i:]})
						if didUseKey {
							lastKeys = NewKey("")
							break
						}
					}

				}
			case *tcell.EventResize:
				w.Display.FullRender()
			}
		}
	}()

	<-w.quit
	w.Display.End()
}
