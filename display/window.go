package display

import "github.com/kiasaki/editing/text"

type WindowKind int

const (
	WindowNode            WindowKind = iota
	WindowVerticalSplit              = iota
	WindowHorizontalSplit            = iota
)

type Window struct {
	// Tree related
	kind   WindowKind
	top    *Window
	bottom *Window
	left   *Window
	right  *Window

	// Node related
	buffer *text.Buffer
}

func WindowNewNode(buffer *text.Buffer) *Window {
	return &Window{
		kind:   WindowNode,
		buffer: buffer,
	}
}
