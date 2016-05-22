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
	if b.GetChar() != '\n' {
		b.PointMove(-1)
	}
}

func moveRight(w *World, b *Buffer, k *Key) {
	if b.Modes.IsEditingModeNamed("normal") {
		// Normal mode must stop before the \n char and can't hover it
		b.PointMove(2)
		if b.GetChar() == '\n' {
			b.PointMove(-1)
		}
		b.PointMove(-1)
	} else {
		b.PointMove(1)
		if b.GetChar() == '\n' {
			b.PointMove(-1)
		}
	}
}

func moveDown(w *World, b *Buffer, k *Key) {
	pointBefore := b.Point
	moveBeginningOfLine(w, b, k)
	column := pointBefore - b.Point

	moveEndOfLine(w, b, k)
	// Go over \n then to 1st char next line
	b.PointMove(2)
	for i := 0; i < int(column); i++ {
		moveRight(w, b, k)
	}
}

func moveUp(w *World, b *Buffer, k *Key) {
	pointBefore := b.Point
	moveBeginningOfLine(w, b, k)
	column := pointBefore - b.Point

	b.PointMove(-1)
	moveBeginningOfLine(w, b, k)
	for i := 0; i < int(column); i++ {
		moveRight(w, b, k)
	}
}

func moveBeginningOfLine(w *World, b *Buffer, k *Key) {
	b.MoveToPreviousChar('\n')
}

func moveEndOfLine(w *World, b *Buffer, k *Key) {
	b.MoveToNextChar('\n')
	if b.Modes.IsEditingModeNamed("normal") {
		b.PointMove(-1)
	}
}

func deleteChar(w *World, b *Buffer, k *Key) {
	b.Delete(1)
}

func init() {
	NormalMode = NewMode("normal", ModeEditing, map[*Key]func(*World, *Buffer, *Key){
		NewKey("i"): func(w *World, b *Buffer, k *Key) {
			b.EnterInsertMode()
		},
		NewKey("a"): func(w *World, b *Buffer, k *Key) {
			b.EnterInsertMode()
			moveRight(w, b, k)
		},
		NewKey("A"): func(w *World, b *Buffer, k *Key) {
			b.EnterInsertMode()
			moveEndOfLine(w, b, k)
		},
		NewKey("h"): moveLeft,
		NewKey("l"): moveRight,
		NewKey("j"): moveDown,
		NewKey("k"): moveUp,
		NewKey("0"): moveBeginningOfLine,
		NewKey("$"): moveEndOfLine,
		NewKey("x"): deleteChar,
	})
	InsertMode = NewMode("insert", ModeEditing, map[*Key]func(*World, *Buffer, *Key){
		NewKey("ESC"): func(w *World, b *Buffer, k *Key) {
			b.PointMove(-1)
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
			b.PointMove(-1)
		},
		NewKey("RIGHT"): func(w *World, b *Buffer, k *Key) {
			b.PointMove(1)
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
