package main

type WindowKind int

const (
	WindowNode            WindowKind = iota
	WindowVerticalSplit              = iota
	WindowHorizontalSplit            = iota
)

type Window struct {
	// Tree related
	Kind   WindowKind
	Top    *Window
	Bottom *Window
	Left   *Window
	Right  *Window

	// Node related
	Buffer         *Buffer
	TopOfWindow    *Mark
	BottomOfWindow *Mark
}

func NewWindowNode(buffer *Buffer) *Window {
	return &Window{
		Kind:   WindowNode,
		Buffer: buffer,
	}
}

func (w *Window) Frame(windowHeight int) {
	/*
		var preferedPercentage float64 = 0.35
		var newScrollLocation int

		buffer := w.Buffer
		markSaved := buffer.MarkCreate()

		if w.TopOfWindow == nil {
			w.TopOfWindow = buffer.MarkCreate()
			w.TopOfWindow.Location = NewLocation(-1)
		}

		buffer.FindFirstInBackward("\n")

		var count int
		for count = 0; count < windowHeight; count++ {
			// If we are at the top of the window
			if buffer.IsPointAtMark(w.TopOfWindow) {
				break
			}

			// If we reached the start of the buffer
			if CompareLocations(buffer.Point, NewLocation(-1)) != LocationAfter {
				break
			}

			if count == int(float64(windowHeight)*preferedPercentage) {
				newScrollLocation = count
			}

			buffer.PointMove(-1)
			buffer.FindFirstInBackward("\n")
		}

		if count >= windowHeight {
			w.TopOfWindow.Location = NewLocation(newScrollLocation)
		}

		buffer.MarkDelete(markSaved)
	*/
}
