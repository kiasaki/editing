package main

type Cursor struct {
	Char   int
	Line   int
	buffer *Buffer
}

func NewCursor(c, l int, b *Buffer) *Cursor {
	return &Cursor{
		Char:   c,
		Line:   l,
		buffer: b,
	}
}

func (c *Cursor) currentLine() string {
	return c.buffer.Lines[c.Line-1]
}

func (c *Cursor) currentLineMaxLength() int {
	line := c.currentLine()
	lineLength := len(line)
	if c.buffer.IsInNormalMode() {
		lineLength = len(line) - 1
	}
	return lineLength
}

func (c *Cursor) Clone() *Cursor {
	return NewCursor(c.Char, c.Line, c.buffer)
}

func (c *Cursor) ToBufferChar() int {
	pos := 0
	for l := 0; l < c.Line-1; l++ {
		pos += len(c.buffer.Lines[l]) + 1 // add one for \n
	}
	pos += c.Char
	return pos
}

func (c *Cursor) SetChar(pos int) {
	lineMaxLength := c.currentLineMaxLength()
	if pos < 0 {
		c.Char = 0
	} else if pos > lineMaxLength {
		c.Char = lineMaxLength
	} else {
		c.Char = pos
	}
}

func (c *Cursor) SetLine(pos int) {
	if pos < 1 {
		c.Line = 1
	} else if pos > len(c.buffer.Lines) {
		c.Line = len(c.buffer.Lines)
	} else {
		c.Line = pos
	}
}

func (c *Cursor) Set(x, y int) {
	c.SetLine(y)
	c.SetChar(x)
}

func (c *Cursor) EnsureValidPosition() {
	c.Set(c.Char, c.Line)
}

func (c *Cursor) MoveChar(posMove int) {
	c.SetChar(c.Char + posMove)
}

func (c *Cursor) MoveLine(posMove int) {
	c.SetLine(c.Line + posMove)
}

func (c *Cursor) Left() {
	c.SetChar(c.Char - 1)
}

func (c *Cursor) Right() {
	c.SetChar(c.Char + 1)
}

func (c *Cursor) Up() {
	c.Set(c.Char, c.Line-1)
}

func (c *Cursor) Down() {
	c.Set(c.Char, c.Line+1)
}

func (c *Cursor) BeginningOfLine() {
	c.SetChar(0)
}

func (c *Cursor) EndOfLine() {
	lineMaxLength := c.currentLineMaxLength()
	c.Char = lineMaxLength
}
