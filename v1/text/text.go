package text

// Char -> Word -> Line -> Page -> Document

type Position struct {
	Line int
	Char int
}

func PositionNew(l int, c int) Position {
	return Position{
		Line: l,
		Char: c,
	}
}

type PositionComparison int

const (
	PositionBefore PositionComparison = iota
	PositionSame                      = iota
	PositionAfter                     = iota
)

func PositionCompare(p1 Position, p2 Position) PositionComparison {
	if p1.Line == p2.Line {
		if p1.Char == p2.Char {
			return PositionSame
		} else if p1.Char < p2.Char {
			return PositionBefore
		} else {
			return PositionAfter
		}
	} else if p1.Line < p2.Line {
		return PositionBefore
	} else {
		return PositionAfter
	}
}

type Mark struct {
	Name     string
	Location Position
	IsFixed  bool
}

type Mode struct {
	Name   string
	Update *func(*World) bool
}
