package main

import (
	"errors"
	"strings"

	"github.com/kiasaki/ry/lang"
)

func NewInterpretor(w *World) (*lang.Interp, error) {
	interpretor := lang.NewInterp()
	preludeReader := strings.NewReader(lang.Prelude)
	if !lang.Run(interpretor, preludeReader) {
		return nil, errors.New("Error parsing lang prelude")
	}
	return interpretor, nil
}

func addEditorBuiltinsToIterpretor(w *World, i *lang.Interp) {
	addEditorBuiltinsForBuffers(w, i)
	addEditorBuiltinsForSettings(w, i)
}

func addEditorBuiltinsForBuffers(w *World, i *lang.Interp) {
	i.Def("buffers", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("current-buffer", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("make-buffer", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("delete-buffer", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("switch-to-buffer", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("buffer-name", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("buffer-path", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("buffer-lines", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("buffer-line-count", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("buffer-cursor", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("buffer-insert", 1, func(a []interface{}) interface{} {
		return nil
	})
	i.Def("buffer-delete", 1, func(a []interface{}) interface{} {
		return nil
	})
}

func addEditorBuiltinsForSettings(w *World, i *lang.Interp) {
	i.Def("get-option", 1, func(a []interface{}) interface{} {
		if _, ok := a[0].(string); !ok {
			panic(lang.NewEvalError("get-option expected name to be a string", a[0]))
		}
		switch value := w.Config.Settings[a[0].(string)].(type) {
		case bool:
			if value == false {
				return nil
			} else {
				return true
			}
		case int:
			return float64(value)
		case string:
			return value
		case float64:
			return value
		default:
			panic(lang.NewEvalError("get-option can't convert setting value to lisp value", value))
		}
	})

	i.Def("set-option", 2, func(a []interface{}) interface{} {
		var value interface{} = nil
		if _, ok := a[0].(string); !ok {
			panic(lang.NewEvalError("set-option expected name to be a string", a[0]))
		}

		if a[1] == true {
			value = true
		}

		switch v := a[1].(type) {
		case string:
			value = v
		case float64:
			value = int(v)
		default:
			panic(lang.NewEvalError("set-option can only take string, number or boolean settings", v))
		}

		w.Config.Settings[a[0].(string)] = value

		return nil
	})
}
