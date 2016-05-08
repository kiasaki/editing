package text

import "time"

type Buffer struct {
	Name string

	FileName   string
	FileTime   time.Time
	IsModified bool

	Point Position
	Marks []Mark
	Modes []Mode

	Contents []string
}

func BufferNew(name string) *Buffer {
	return &Buffer{
		Name: name,

		FileName:   "",
		FileTime:   time.Now(),
		IsModified: false,

		Point: PositionNew(1, -1),
		Marks: []Mark{},
		Modes: []Mode{},

		Contents: []string{""},
	}
}

func (b *Buffer) Clear() {
	b.Marks = []Mark{}
	b.Contents = []string{}
}

func (b *Buffer) LoadFile() {
	panic("Not implemented")
}

func (b *Buffer) SaveFile() {
	panic("Not implemented")
}

func (b *Buffer) Start(p Position) {
	b.Point = PositionNew(1, -1)
}

func (b *Buffer) End(p Position) {
	lastLine := b.Contents[len(b.Contents)-1]
	b.Point = PositionNew(len(b.Contents), len(lastLine)-1)
}

func (b *Buffer) Move(count int) {
	b.Point = b.PositionMove(b.Point, count)
}

func (b *Buffer) PositionMove(pos Position, count int) Position {
	forward := true
	if count < 0 {
		count = 0 - count
		forward = false
	}

	for i := 0; i < count; i++ {
		if forward {
			// End of the line is len -1 (array normal) -1 (char is to the left)
			if pos.Char == len(b.Contents[pos.Line-1])-2 {
				// We where at the EOL, go to next line
				// (if not at the EOF)
				if pos.Line < len(b.Contents) {
					pos.Line++
					pos.Char = -1
				}
			} else {
				// We advance one char
				pos.Char++
			}
		} else {
			if pos.Char == -1 {
				// We where at the EOL, go to previous line
				// (if not on first line)
				if pos.Line > 1 {
					pos.Line--
					pos.Char = len(b.Contents[pos.Line-1]) - 2
				}
			} else {
				// We go back one char
				pos.Char--
			}
		}
	}
	return pos
}

func (b *Buffer) PositionToCount(p Position) int {
	count := 0

	// -1 because lines start at one but Contents still is a 0 based array
	for l := 1; l < p.Line && l-1 < len(b.Contents); l++ {
		count += len(b.Contents[l-1])
	}

	// +1 as we are always to the left and in between with chars
	count += p.Char + 1

	return count
}

func (b *Buffer) CountToPosition(count int) Position {
	pos := PositionNew(1, -1)
	pos = b.PositionMove(pos, count)
	return pos
}
