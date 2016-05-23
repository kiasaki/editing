package main

import (
	"errors"
	"strings"

	"github.com/kiasaki/ry/lang"
)

func NewInterpretor() (*lang.Interp, error) {
	interpretor := lang.NewInterp()
	preludeReader := strings.NewReader(lang.Prelude)
	if !lang.Run(interpretor, preludeReader) {
		return nil, errors.New("Error parsing lang prelude")
	}
	return interpretor, nil
}
