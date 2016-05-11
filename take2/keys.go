package main

import (
	"strings"

	"github.com/gdamore/tcell"
)

type KeyStroke struct {
	modMask tcell.ModMask
	key     tcell.Key
	rune    rune
}

func NewKeyStroke(representation string) *KeyStroke {
	parts := strings.Split(representation, "-")

	// Modifiers
	modMask := tcell.ModNone
	for _, part := range parts[:len(parts)-1] {
		switch part {
		case "C":
			modMask |= tcell.ModCtrl
		case "S":
			modMask |= tcell.ModShift
		case "A":
			modMask |= tcell.ModAlt
		case "M":
			modMask |= tcell.ModMeta
		}
	}

	// Key
	var r rune
	var key tcell.Key
	lastPart := parts[len(parts)-1]
	switch lastPart {
	case "DEL":
		key = tcell.KeyBackspace2
	case "RET":
		key = tcell.KeyEnter
	case "SPC":
		key = tcell.KeySpace
	case "ESC":
		key = tcell.KeyEscape
	case "TAB":
		key = tcell.KeyTab
	default:
		key = tcell.KeyRune
		r = []rune(lastPart)[0]
	}

	return &KeyStroke{modMask, key, r}
}

func NewKeyStrokeFromKeyEvent(ev *tcell.EventKey) *KeyStroke {
	return &KeyStroke{
		modMask: ev.Modifiers(),
		key:     ev.Key(),
		rune:    ev.Rune(),
	}
}

func (k1 *KeyStroke) Matches(k2 *KeyStroke) bool {
	masksMatch := k1.modMask == k2.modMask
	keysMatch := k1.key == k2.key

	runesMatch := true
	if k1.key == tcell.KeyRune {
		runesMatch = k1.rune == k2.rune
	}

	return masksMatch && keysMatch && runesMatch
}

type Key struct {
	keys []*KeyStroke
}

func NewKey(representation string) *Key {
	keys := []*KeyStroke{}

	individualKeyStrokes := strings.Split(representation, " ")
	for _, keyStrokeRep := range individualKeyStrokes {
		if len(keyStrokeRep) > 0 {
			keys = append(keys, NewKeyStroke(keyStrokeRep))
		}
	}

	return &Key{keys}
}

func (k1 *Key) Matches(k2 *Key) bool {
	if len(k1.keys) != len(k2.keys) {
		return false
	}

	for i := 0; i < len(k1.keys); i++ {
		if !k1.keys[i].Matches(k2.keys[i]) {
			return false
		}
	}

	return true
}