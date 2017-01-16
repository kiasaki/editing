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

func moveBeginingOfBuffer(w *World, b *Buffer, k *Key) {
	b.Cursor.SetLine(0)
	b.Cursor.SetChar(0)
}

func moveEndOfBuffer(w *World, b *Buffer, k *Key) {
	b.Cursor.SetLine(len(b.Lines))
	b.Cursor.SetChar(0)
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
		NewKey("h"):   moveLeft,
		NewKey("l"):   moveRight,
		NewKey("j"):   moveDown,
		NewKey("k"):   moveUp,
		NewKey("0"):   moveBeginningOfLine,
		NewKey("$"):   moveEndOfLine,
		NewKey("x"):   deleteChar,
		NewKey("g g"): moveBeginingOfBuffer,
		NewKey("G"):   moveEndOfBuffer,
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
