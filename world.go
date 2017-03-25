package main

import (
	"github.com/gdamore/tcell"
)

type World struct {
	lastKeys    *Key
	quit        chan (struct{})
	Config      *Config
	Display     *Display
	Buffers     []*Buffer
	Interpretor interface{}
	Command     string
	Message     string
	MessageType string
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

	w.Display, err = NewDisplay(w.Config, w)
	if err != nil {
		return err
	}

	w.Interpretor, err = NewInterpretor(w)
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
		w.lastKeys = NewKey("")

		for {
			if len(w.Buffers) == 0 {
				w.Quit()
			}

			// Now wait for and handle user event
			select {
			case ev := <-terminalEventChan:
				switch ev := ev.(type) {
				case *tcell.EventKey:
					if ev.Key() == tcell.KeyCtrlQ {
						// TODO remove safeguard quit once bindings
						// implentation works well
						w.Quit()
						return
					} else if ev.Key() == tcell.KeyCtrlG {
						// Cancel entered keys
						w.lastKeys = NewKey("")
					} else if ev.Key() == tcell.KeyEscape && w.lastKeys.Length() > 0 {
						// Cancel entered keys
						w.lastKeys = NewKey("")
					} else {
						// Add lastest key store to what was already types and
						// check if a binding will handle it.
						// If used up, reset types keys
						keyStroke := NewKeyStrokeFromKeyEvent(ev)
						w.lastKeys.AppendKeyStroke(keyStroke)

						// Loop on the set of keys last entered and see if it matches any
						// bound function for the current mode
						// "C-a b c M-x" might not be bound, neither is "b c M-x" but if
						// get to just the last key "M-x" is bound.
						for i := 0; i < len(w.lastKeys.keys); i++ {
							didUseKey := w.Display.HandleEvent(w, &Key{w.lastKeys.keys[i:]})
							if didUseKey {
								w.lastKeys = NewKey("")
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
