package main

import "github.com/gdamore/tcell"

var CatchAllKey *Key = NewKey("")

type ModeKind string

const (
	ModeEditing ModeKind = "ModeEditing"
	ModeMajor            = "ModeMajor"
	ModeMinor            = "ModeMinor"
)

type Mode struct {
	Name     string
	Kind     ModeKind
	Commands map[*Key]func(*World, *Buffer, *Key)
}

func (m *Mode) HandleEvent(w *World, b *Buffer, key *Key) bool {
	var catchAll func(*World, *Buffer, *Key) = nil

	for cKey, cFn := range m.Commands {
		// Save catch all as last
		if cKey == CatchAllKey {
			catchAll = cFn
		}
		if cKey.Matches(key) {
			cFn(w, b, key)
			return true
		}
	}

	// If we had a catchall key, run it's function
	if catchAll != nil {
		catchAll(w, b, key)
		return true
	}

	return false
}

func NewMode(name string, kind ModeKind, commands map[*Key]func(*World, *Buffer, *Key)) *Mode {
	return &Mode{
		Name:     name,
		Kind:     kind,
		Commands: commands,
	}
}

var NormalMode *Mode
var InsertMode *Mode
var ReplaceMode *Mode
var VisualMode *Mode
var VisualLineMode *Mode

func moveLeft(w *World, b *Buffer, k *Key) {
	b.Cursor.Left()
}

func moveRight(w *World, b *Buffer, k *Key) {
	b.Cursor.Right()
}

func moveUp(w *World, b *Buffer, k *Key) {
	b.Cursor.Up()
}

func moveDown(w *World, b *Buffer, k *Key) {
	b.Cursor.Down()
}

func moveBeginningOfLine(w *World, b *Buffer, k *Key) {
	b.Cursor.BeginningOfLine()
}

func moveEndOfLine(w *World, b *Buffer, k *Key) {
	b.Cursor.EndOfLine()
}

func deleteChar(w *World, b *Buffer, k *Key) {
	b.Delete(1)
}

func deleteLine(w *World, b *Buffer, k *Key) {
	char := b.Cursor.Char

	b.Cursor.SetChar(0)
	start := b.Cursor.ToBufferChar()
	b.Cursor.EndOfLine()
	end := b.Cursor.ToBufferChar()
	b.Cursor.SetChar(0)
	b.Delete((end + 2) - start)

	b.Cursor.SetChar(char)
}

func moveBeginingOfBuffer(w *World, b *Buffer, k *Key) {
	b.Cursor.SetLine(0)
	b.Cursor.SetChar(0)
}

func moveEndOfBuffer(w *World, b *Buffer, k *Key) {
	b.Cursor.SetLine(len(b.Lines))
	b.Cursor.SetChar(0)
}

func windowSplitHorizontally(w *World, b *Buffer, k *Key) {
	newWindow := NewWindowNode(w.Display.CurrentBuffer())
	w.Display.ReplaceCurrentWindow(func(w *Window) *Window {
		return &Window{
			Kind:   WindowHorizontalSplit,
			Top:    w,
			Bottom: newWindow,
		}
	})
	w.Display.SetCurrentWindow(newWindow)
}

func windowSplitVertically(w *World, b *Buffer, k *Key) {
	newWindow := NewWindowNode(w.Display.CurrentBuffer())
	w.Display.ReplaceCurrentWindow(func(w *Window) *Window {
		return &Window{
			Kind:  WindowVerticalSplit,
			Left:  w,
			Right: newWindow,
		}
	})
	w.Display.SetCurrentWindow(newWindow)
}

func invertDirection(going string) string {
	switch going {
	case "top":
		return "bottom"
	case "bottom":
		return "top"
	case "left":
		return "right"
	case "right":
		return "left"
	}
	panic("unreachable")
}

func windowMove(w *World, going string, nParent int) {
	w.Display.WindowTree.EnsureParents(nil)
	window := w.Display.CurrentWindow
	changedDirection := false

	var searchingForKind WindowKind
	if going == "left" || going == "right" {
		searchingForKind = WindowVerticalSplit
	} else {
		searchingForKind = WindowHorizontalSplit
	}

	// dig up
	for window != nil && window.Kind != searchingForKind {
		window = window.Parent
	}
	// go up some more if we need to
	for i := nParent; i > 0 && window != nil; i-- {
		window = window.Parent
	}
	if window == nil {
		// no where to move left
		return
	}

	// dig down
	for window.Kind != WindowNode {
		switch window.Kind {
		case WindowHorizontalSplit:
			if going == "top" {
				window = window.Top
			} else {
				window = window.Bottom
			}
		case WindowVerticalSplit:
			if going == "left" {
				window = window.Left
			} else {
				window = window.Right
			}
		}
		if !changedDirection {
			changedDirection = true
			going = invertDirection(going)
		}
	}

	// update current window with first leaf found
	if w.Display.CurrentWindow == window && nParent < 10 {
		windowMove(w, invertDirection(going), nParent+1)
		return
	}
	w.Display.SetCurrentWindow(window)
}

func windowMoveLeft(w *World, b *Buffer, k *Key) {
	windowMove(w, "left", 0)
}
func windowMoveRight(w *World, b *Buffer, k *Key) {
	windowMove(w, "right", 0)
}
func windowMoveUp(w *World, b *Buffer, k *Key) {
	windowMove(w, "top", 0)
}
func windowMoveDown(w *World, b *Buffer, k *Key) {
	windowMove(w, "bottom", 0)
}

func init() {
	NormalMode = NewMode("normal", ModeEditing, map[*Key]func(*World, *Buffer, *Key){
		NewKey("i"): func(w *World, b *Buffer, k *Key) {
			b.EnterInsertMode()
		},
		NewKey("I"): func(w *World, b *Buffer, k *Key) {
			b.Cursor.BeginningOfLine()
			b.EnterInsertMode()
		},
		NewKey("a"): func(w *World, b *Buffer, k *Key) {
			b.EnterInsertMode()
			b.Cursor.Right()
		},
		NewKey("A"): func(w *World, b *Buffer, k *Key) {
			b.EnterInsertMode()
			b.Cursor.EndOfLine()
		},
		NewKey("o"): func(w *World, b *Buffer, k *Key) {
			b.EnterInsertMode()
			b.Cursor.EndOfLine()
			b.Insert("\n")
			b.Cursor.Down()
		},
		NewKey("O"): func(w *World, b *Buffer, k *Key) {
			b.Cursor.BeginningOfLine()
			b.Insert("\n")
			b.EnterInsertMode()
		},
		NewKey("h"):   moveLeft,
		NewKey("l"):   moveRight,
		NewKey("j"):   moveDown,
		NewKey("k"):   moveUp,
		NewKey("0"):   moveBeginningOfLine,
		NewKey("$"):   moveEndOfLine,
		NewKey("x"):   deleteChar,
		NewKey("d d"): deleteLine,
		NewKey("g g"): moveBeginingOfBuffer,
		NewKey("G"):   moveEndOfBuffer,

		NewKey("C-w s"): windowSplitHorizontally,
		NewKey("C-w v"): windowSplitVertically,
		NewKey("C-w h"): windowMoveLeft,
		NewKey("C-w j"): windowMoveDown,
		NewKey("C-w k"): windowMoveUp,
		NewKey("C-w l"): windowMoveRight,
		NewKey("C-h"):   windowMoveLeft,
		NewKey("C-j"):   windowMoveDown,
		NewKey("C-k"):   windowMoveUp,
		NewKey("C-l"):   windowMoveRight,
	})
	InsertMode = NewMode("insert", ModeEditing, map[*Key]func(*World, *Buffer, *Key){
		NewKey("ESC"): func(w *World, b *Buffer, k *Key) {
			moveLeft(w, b, k)
			b.EnterNormalMode()
		},
		NewKey("RET"): func(w *World, b *Buffer, k *Key) {
			b.NewLineAndIndent()
		},
		NewKey("BAK"): func(w *World, b *Buffer, k *Key) {
			b.Backspace()
		},
		NewKey("BAK2"): func(w *World, b *Buffer, k *Key) {
			b.Backspace()
		},
		NewKey("DEL"): deleteChar,
		NewKey("SPC"): func(w *World, b *Buffer, k *Key) {
			b.Insert(" ")
		},
		NewKey("TAB"): func(w *World, b *Buffer, k *Key) {
			if tabToSpaces, ok := w.Config.GetSetting("tabtospaces"); ok && tabToSpaces.(bool) {
				tabWidth := 4
				tabWidthSetting, ok := w.Config.GetSetting("tabwidth")
				if ok {
					tabWidth = tabWidthSetting.(int)
				}
				b.Insert(Pad("", tabWidth, ' '))
			} else {
				b.Insert("\t")
			}
		},
		NewKey("LEFT"): func(w *World, b *Buffer, k *Key) {
			moveLeft(w, b, k)
		},
		NewKey("RIGHT"): func(w *World, b *Buffer, k *Key) {
			moveRight(w, b, k)
		},
		// Make sure catch all stays last so it doesn't hide other keys
		CatchAllKey: func(w *World, b *Buffer, k *Key) {
			lastKeyStroke := k.keys[len(k.keys)-1]
			if lastKeyStroke.key == tcell.KeyRune {
				b.Insert(string(lastKeyStroke.rune))
			}
		},
	})
}
