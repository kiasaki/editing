package main

import "strings"

func NewInterpretor(w *World) (interface{}, error) {
	return nil, nil
}

func (w *World) ExecuteCommand() {
	commandParts := strings.Split(w.Command[1:], " ")
	if len(commandParts) == 0 {
		return
	}

	name := commandParts[0]
	args := commandParts[1:]

	if name == "w" || name == "write" {
		commandWrite(w, args)
	} else if name == "q" || name == "quit" {
		commandQuit(w, args)
	} else if name == "wq" {
		commandWrite(w, args)
		commandQuit(w, args)
	}
}

func commandWrite(w *World, args []string) {
	// TODO
}
func commandQuit(w *World, args []string) {
	// TODO check for unsaved changes
	w.Display.CurrentBuffer().Close(w)
}
