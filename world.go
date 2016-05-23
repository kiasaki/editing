package main

import (
	"github.com/gdamore/tcell"
	"github.com/kiasaki/ry/lang"
)

type World struct {
	quit        chan (struct{})
	Config      *Config
	Display     *Display
	Buffers     []*Buffer
	Interpretor *lang.Interp
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

	w.Interpretor, err = NewInterpretor()
	return err
}

func (w *World) Quit() {
	if w.quit != nil {
		close(w.quit)
	}
}

func (w *World) Run() {
	w.quit = make(chan struct{})

	terminalEventChan := make(chan tcell.Event, 20)
	go func() {
		for {
			terminalEventChan <- w.Display.Screen.PollEvent()
		}
	}()

	go func() {
		defer handlePanics()

		lastKeys := NewKey("")

		for {
			// TODO load up files in args
			if len(w.Buffers) == 0 {
				scratchBuffer := NewBuffer("*scratch*", "")
				w.Buffers = append(w.Buffers, scratchBuffer)
				w.Display.WindowTree = NewWindowNode(scratchBuffer)
				w.Display.SetCurrentWindow(w.Display.WindowTree)
			}

			// Now wait for and handle user event
			select {
			case ev := <-terminalEventChan:
				switch ev := ev.(type) {
				case *tcell.EventKey:
					// DEBUGING
					w.Display.write(tcell.StyleDefault, 5, 20, lastKeys.String())
					w.Display.Screen.Show()

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
					// TODO recompute window sizes
				}
				w.Display.Render()
			}
		}
	}()

	<-w.quit
	w.Display.End()
}
