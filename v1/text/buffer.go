package text

import (
	"time"
	"unicode/utf8"
)

type Buffer struct {
	name string

	fileName   string
	fileTime   time.Time
	isModified bool

	point Position
	marks []Mark
	modes []Mode

	contents []string
}

func BufferNew(name string) *Buffer {
	return &Buffer{
		name: name,

		fileName:   "",
		fileTime:   time.Now(),
		isModified: false,

		point: PositionNew(1, -1),
		marks: []Mark{},
		modes: []Mode{},

		contents: []string{""},
	}
}

func (b *Buffer) Name() string {
	return b.name
}

func (b *Buffer) SetName(name string) {
	b.name = name
}

func (b *Buffer) Contents() []string {
	return b.contents
}

func (b *Buffer) Point() Position {
	return b.point
}

func (b *Buffer) Clear() {
	b.marks = []Mark{}
	b.contents = []string{}
}

func (b *Buffer) LoadFile() {
	panic("Not implemented")
}

func (b *Buffer) SaveFile() {
	panic("Not implemented")
}

func (b *Buffer) Start(p Position) {
	b.point = PositionNew(1, -1)
}

func (b *Buffer) End(p Position) {
	lastLine := b.contents[len(b.contents)-1]
	b.point = PositionNew(len(b.contents), len(lastLine)-1)
}

func (b *Buffer) Move(count int) {
	b.point = b.PositionMove(b.point, count)
}

func (b *Buffer) PositionMove(pos Position, count int) Position {
	forward := true
	if count < 0 {
		count = 0 - count
		forward = false
	}

	for i := 0; i < count; i++ {
		if forward {
			// End of the line is len...
			//   -1 array normal
			//   -1 char representation is to the left
			//   +1 for new line char
			if pos.Char == utf8.RuneCountInString(b.contents[pos.Line-1])-1 {
				// We where at the EOL, go to next line
				// (if not at the EOF)
				if pos.Line < len(b.contents) {
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
					pos.Char = len(b.contents[pos.Line-1]) - 2
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

	// -1 because lines start at one but contents still is a 0 based array
	for l := 1; l < p.Line && l-1 < len(b.contents); l++ {
		count += len(b.contents[l-1])
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

// mark_add
// mark_delete
// mark_to_point
// point_to_mark
// set_mark_position
// get_mark_position
// is_point_at_mark
// is_point_before_mark
// is_point_after_mark
// swap_point_and_mark

func (b *Buffer) GetChar() rune {
	for i, ch := range b.contents[b.point.Line-1] {
		if i == b.point.Char+1 {
			return ch
		}
	}

	// If we didn't find the point's index in the current line
	// we at the EOL, return the new line char
	return '\n'
}

func (b *Buffer) GetString(count int) string {
	str := ""
	l := b.point.Line - 1
	c := b.point.Char + 1

	for i := count; i > 0; {
		// Take chars while we till the end of the line
		for j, ch := range b.contents[l] {
			if i == 0 {
				break
			}
			if j > c {
				c++
				str += string(ch)
				i--
			}
		}
		// No next line to skip to
		if len(b.contents) == l+1 {
			break
		}
		// Move to next line
		l++
		c = -1
		i--
		str += "\n"
	}

	return str
}

func (b *Buffer) LineCount() int {
	return len(b.contents)
}
func (b *Buffer) CharCount() int {
	count := 0
	for _, line := range b.contents {
		// Get an accurate rune count (not byte) by ranging over line string
		for _ = range line {
			count++
		}
	}
	return count
}

func (b *Buffer) InsertChar(ch rune) {
	b.InsertString(string(ch))
}

func (b *Buffer) InsertString(str string) {
	l := b.point.Line - 1
	c := b.point.Char + 1
	line := b.contents[l]
	b.contents[l] = ""

	if len(line) == 0 {
		b.contents[l] = str
		b.Move(utf8.RuneCountInString(str))
		return
	}
	for i, ch := range line {
		// Insert string at point
		if i == c {
			b.contents[l] += str
		}
		b.contents[l] += string(ch)
	}

	b.Move(utf8.RuneCountInString(str))
}

func (b *Buffer) ReplaceChar(ch rune) {
	b.ReplaceString(string(ch))
}

func (b *Buffer) ReplaceString(str string) {
	// TODO make UTF8 aware
	l := b.point.Line - 1
	c := b.point.Char + 1
	line := b.contents[l]
	b.contents[l] = line[:c] + str

	// We're cutting len(str) further off the 2nd half of the line,
	// check that the line is long enough for that
	if len(line) > c+len(str) {
		b.contents[l] += line[c+len(str):]
	}
}

func (b *Buffer) Delete(count int) {
	l := b.point.Line - 1
	c := b.point.Char + 1
	line := b.contents[l]
	rest := line[c:]
	b.contents[l] = line[:c]

	for i := count; i > 0; {
		if len(rest) > 0 {
			rest = rest[1:]
			count--
			continue
		}

		// We are the the end of the line, move to next
		if !(len(b.contents) > l+1) {
			break
		}
		l++
		count--
		rest = b.contents[l]
	}

	// Add what we have left back to the current line
	b.contents[l] += rest
}

func (b *Buffer) DeleteRegion(markName string) {
	panic("not implemented")
}

func (b *Buffer) CopyRegion(markName string) string {
	panic("not implemented")
}
