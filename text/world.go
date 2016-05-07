package text

var currentWorld *World

type World struct {
	Buffers       []*Buffer
	CurrentBuffer *Buffer
}

func WorldNew() *World {
	scratchBuffer := BufferNew("*scratch*")
	return &World{
		Buffers:       []*Buffer{scratchBuffer},
		CurrentBuffer: scratchBuffer,
	}
}

func (w *World) Init() error {
	currentWorld = WorldNew()
	return nil
}

func (w *World) End() error {
	panic("Not implemented")
}

func (w *World) SetCurrentBuffer(name string) bool {
	for _, buffer := range w.Buffers {
		if buffer.Name == name {
			w.CurrentBuffer = buffer
			return true
		}
	}

	return false
}

func (w *World) DeleteBuffer(name string) bool {
	for i, buffer := range w.Buffers {
		if buffer.Name == name {
			w.Buffers = append(w.Buffers[:i], w.Buffers[i+1:]...)

			if w.CurrentBuffer == buffer {
				if len(w.Buffers) > 0 {
					// TODO: Be smart about buffer usage history
					// and use last in use buffer
					w.CurrentBuffer = w.Buffers[0]
				} else {
					// No buffers where left, re-create the scratch buffer
					scratchBuffer := BufferNew("*scratch*")
					w.Buffers = []*Buffer{scratchBuffer}
					w.CurrentBuffer = scratchBuffer
				}
			}

			return true
		}
	}

	return false
}
