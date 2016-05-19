package main

import (
	"crypto/md5"
	"strings"

	"github.com/kiasaki/go-rope"
)

type Location int

func NewLocation(l int) Location {
	return Location(l)
}

type Mark struct {
	Location Location
	Fixed    bool
}

func NewMark(loc Location, fixed bool) *Mark {
	return &Mark{Location: loc, Fixed: fixed}
}

type Buffer struct {
	r *rope.Rope

	Name        string
	Path        string
	LastSaveSum [16]byte
	Modified    bool
	Point       Location
	Marks       []*Mark

	Modes *ModeList

	Lines     []string
	LineCount int
}

func NewBuffer(name, text string) *Buffer {
	buffer := &Buffer{
		r: rope.New(text),

		Name:     name,
		Path:     name,
		Modified: false,
		Point:    NewLocation(-1),
		Marks:    []*Mark{},

		Modes: NewModeList(),
	}

	buffer.LastSaveSum = md5.Sum([]byte(text))

	buffer.CacheLines()

	buffer.EnterNormalMode()

	return buffer
}

func (b *Buffer) CacheLines() {
	b.Lines = strings.Split(b.String(), "\n")
	b.LineCount = len(b.Lines)
}

func (b *Buffer) Insert(text string) {
	b.Modified = true
	b.r = b.r.Insert(int(b.Point)+1, text)
	b.CacheLines()
	b.PointMove(len([]rune(text)))
}

func (b *Buffer) Delete(count int) {
	if int(b.Point) < b.r.Len()-1 {
		b.Modified = true
		b.r = b.r.Delete(int(b.Point)+1, count)
		b.CacheLines()
	}
}

func (b *Buffer) Backspace() {
	if b.Point > -1 {
		b.Delete(1)
		b.PointMove(-1)
	}
}

func (b *Buffer) NewLineAndIndent() {
	b.Insert("\n")
}

func (b *Buffer) String() string {
	if b.r.Len() == 0 {
		return ""
	}
	return b.r.String()
}

func (b *Buffer) GetChar() rune {
	if int(b.Point)+2 < b.r.Len()-1 {
		return b.r.Index(int(b.Point) + 1)
	} else {
		return '\x00'
	}
}

func (b *Buffer) MoveToPreviousChar(ch rune) {
	for ; b.Point+1 > 0; b.Point-- {
		if ch == b.GetChar() {
			return
		}
	}
}

func (b *Buffer) MoveToNextChar(ch rune) {
	length := b.r.Len()
	for ; int(b.Point)+1 < length-1; b.Point++ {
		if ch == b.GetChar() {
			return
		}
	}
}

func (b *Buffer) FindFirstInBackward(search string) {
	contents := []rune(b.String())

	// As we are adding charaters backwards to contentsToSearchIn flip "search"
	search = ReverseString(search)
	contentsToSearchIn := ""

	if int(b.Point)+1 >= len(contents) {
		return
	}
	for pos := int(b.Point); pos > 0; pos-- {
		contentsToSearchIn += string(contents[pos+1])
		if strings.HasSuffix(contentsToSearchIn, search) {
			b.Point = NewLocation(pos)
			return
		}
	}
}

func (b *Buffer) FindFirstInForward(search string) {
	contents := []rune(b.String())
	contentsToSearchIn := ""

	for pos := int(b.Point); pos+1 < len(contents); pos++ {
		contentsToSearchIn += string(contents[pos+1])
		if strings.HasSuffix(contentsToSearchIn, search) {
			b.Point = NewLocation(pos)
		}
	}
}

func (b *Buffer) PointMove(amount int) {
	newPoint := int(b.Point) + amount
	if newPoint < 0 {
		b.Point = NewLocation(-1)
	} else if newPoint > b.r.Len()-1 {
		b.Point = NewLocation(b.r.Len() - 1)
	} else {
		b.Point = NewLocation(newPoint)
	}
}

func (b *Buffer) IsPointAtMark(mark *Mark) bool {
	return b.Point == mark.Location
}

func (b *Buffer) MarkCreate() *Mark {
	mark := NewMark(b.Point, false)
	b.Marks = append(b.Marks, mark)
	return mark
}

func (b *Buffer) MarkDelete(mark *Mark) {
	marks := b.Marks
	b.Marks = marks[:0]
	for _, m := range marks {
		if m != mark {
			b.Marks = append(b.Marks, m)
		}
	}
}

func (b *Buffer) ModeAdd(m *Mode) {
	b.Modes.Add(m)
}
func (b *Buffer) EnterNormalMode() {
	b.ModeAdd(NormalMode)
}
func (b *Buffer) EnterInsertMode() {
	b.ModeAdd(InsertMode)
}
func (b *Buffer) EnterReplaceMode() {
	b.ModeAdd(ReplaceMode)
}
func (b *Buffer) EnterVisualMode() {
	b.ModeAdd(VisualMode)
}
func (b *Buffer) EnterVisualLineMode() {
	b.ModeAdd(VisualLineMode)
}

func (b *Buffer) HandleEvent(w *World, key *Key) bool {
	return b.Modes.HandleEvent(w, b, key)
}
