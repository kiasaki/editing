package main

import "github.com/gdamore/tcell"

type ModeKind string

const (
	ModeEditing ModeKind = "ModeEditing"
	ModeMajor            = "ModeMajor"
	ModeMinor            = "ModeMinor"
)

type Mode struct {
	Name     string
	Kind     ModeKind
	Commands map[*Key]func(w *World, b *Buffer)
}

func (m *Mode) HandleEvent(ev *tcell.EventKey) {
}

func NewMode(name string, kind ModeKind, commands map[*Key]func(w *World, b *Buffer)) *Mode {
	return &Mode{
		Name:     name,
		Kind:     kind,
		Commands: commands,
	}
}

var NormalMode Mode
var InsertMode Mode
var ReplaceMode Mode
var VisualMode Mode
var VisualLineMode Mode

func init() {

}
