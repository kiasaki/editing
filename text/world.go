package text

type World struct {
	Buffers       []*Buffer
	currentBuffer *Buffer
}

func WorldNew() *World {
	return &World{
		Buffers:       []*Buffer{},
		currentBuffer: nil,
	}
}

func (w *World) Init() error {
	_, scratchBuffer := w.CreateBuffer("*scratch*")
	w.currentBuffer = scratchBuffer
	return nil
}

func (w *World) End() error {
	return nil
}

func (w *World) CreateBuffer(name string) (string, *Buffer) {
	buffer := BufferNew("*scratch*")
	w.Buffers = append(w.Buffers, buffer)
	return name, buffer
}

func (w *World) CurrentBuffer() *Buffer {
	return w.currentBuffer
}

func (w *World) SetCurrentBuffer(name string) bool {
	for _, buffer := range w.Buffers {
		if buffer.Name() == name {
			w.currentBuffer = buffer
			return true
		}
	}

	return false
}

func (w *World) DeleteBuffer(name string) bool {
	for i, buffer := range w.Buffers {
		if buffer.Name() == name {
			w.Buffers = append(w.Buffers[:i], w.Buffers[i+1:]...)

			if w.currentBuffer == buffer {
				if len(w.Buffers) > 0 {
					// TODO: Be smart about buffer usage history
					// and use last in use buffer
					w.currentBuffer = w.Buffers[0]
				} else {
					// No buffers where left, re-create the scratch buffer
					scratchBuffer := BufferNew("*scratch*")
					w.Buffers = []*Buffer{scratchBuffer}
					w.currentBuffer = scratchBuffer
				}
			}

			return true
		}
	}

	return false
}
