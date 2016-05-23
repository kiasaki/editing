package main

import (
	"crypto/md5"
	"strings"

	"github.com/kiasaki/go-rope"
)

type Mark struct {
	Point *Cursor
	Fixed bool
}

func NewMark(cur *Cursor, fixed bool) *Mark {
	return &Mark{Point: cur, Fixed: fixed}
}

type Buffer struct {
	r *rope.Rope

	Name        string
	Path        string
	LastSaveSum [16]byte
	Modified    bool
	Cursor      *Cursor
	Marks       []*Mark

	Modes *ModeList

	Lines []string
}

func NewBuffer(name, text string) *Buffer {
	buffer := &Buffer{
		r: rope.New(text),

		Name:     name,
		Path:     name,
		Modified: false,
		Marks:    []*Mark{},

		Modes: NewModeList(),
	}

	buffer.Cursor = NewCursor(0, 1, buffer)

	buffer.LastSaveSum = md5.Sum([]byte(text))

	buffer.CacheLines()

	buffer.EnterNormalMode()

	return buffer
}

func (b *Buffer) CacheLines() {
	b.Lines = strings.Split(b.String(), "\n")
}

func (b *Buffer) Insert(text string) {
	b.Modified = true
	b.r = b.r.Insert(b.Cursor.ToBufferChar(), text)
	b.CacheLines()
	b.Cursor.MoveChar(len([]rune(text)))
}

func (b *Buffer) Delete(count int) string {
	start := b.Cursor.ToBufferChar()
	end := start + count

	if end > b.r.Len() {
		end = b.r.Len()
	}

	// Nothing to do here
	if start == end {
		return ""
	}

	b.Modified = true
	removed := b.r.Substr(start+1, end-start).String()
	b.r = b.r.Delete(start+1, end-start)
	b.CacheLines()
	b.Cursor.EnsureValidPosition()
	return removed
}

func (b *Buffer) Backspace() {
	if b.Cursor.Char > 0 {
		b.Delete(1)
		b.Cursor.Left()
	}
}

// TODO implement basic indentation
func (b *Buffer) NewLineAndIndent() {
	b.Insert("\n")
	b.Cursor.Down()
	b.Cursor.BeginningOfLine()
}

func (b *Buffer) String() string {
	if b.r.Len() == 0 {
		return ""
	}
	return b.r.String()
}

func (b *Buffer) IsPointAtMark(mark *Mark) bool {
	return b.Cursor.Line == mark.Point.Line && b.Cursor.Char == mark.Point.Char
}

func (b *Buffer) MarkCreate(fixed bool) *Mark {
	point := b.Cursor.Clone()
	mark := NewMark(point, fixed)
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

func (b *Buffer) IsInNormalMode() bool {
	return b.Modes.IsEditingModeNamed("normal")
}
