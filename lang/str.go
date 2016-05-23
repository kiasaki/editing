package lang

import (
	"fmt"
	"strconv"
	"strings"
)

// Str returns a textual representation of any Lisp expression x.
func Str(x interface{}) string {
	return Str2(x, true)
}

// Str2 returns a textual representation of any Lisp expression x.
// If quoteString is true, any strings in the expression are represented
// with enclosing quotes respectively.
func Str2(x interface{}, quoteString bool) string {
	return str4(x, quoteString, -1, nil)
}

// quotes is a mapping from a quote symbol to its string representation.
var quotes = map[*Sym]string{
	QuoteSym:           "'",
	QuasiquoteSym:      "`",
	UnquoteSym:         ",",
	UnquoteSplicingSym: ",@",
}

func str4(a interface{}, quoteString bool, count int,
	printed map[*Cell]bool) string {
	if a == true {
		return "t"
	}
	switch x := a.(type) {
	case *Cell:
		if x == Nil {
			return "nil"
		}
		if s, ok := x.Car.(*Sym); ok {
			if q, ok := quotes[s]; ok {
				if d, ok := x.Cdr.(*Cell); ok {
					if d.Cdr == Nil {
						return q + str4(d.Car, true, count, printed)
					}
				}
			}
		}
		return "(" + strListBody(x, count, printed) + ")"
	case string:
		if quoteString {
			return strconv.Quote(x)
		}
		return x
	case []interface{}:
		s := make([]string, len(x))
		for i, e := range x {
			s[i] = str4(e, true, count, printed)
		}
		return "[" + strings.Join(s, ", ") + "]"
	case *Sym:
		if x.IsInterned() {
			return x.Name
		}
		return "#:" + x.Name
	}
	return fmt.Sprintf("%v", a)
}

// strListBody makes a string representation of a list, omitting its parens.
func strListBody(x *Cell, count int, printed map[*Cell]bool) string {
	if printed == nil {
		printed = make(map[*Cell]bool)
	}
	if count < 0 {
		count = 4 // threshold of ellipsis for circular lists
	}
	s := make([]string, 0, 10)
	y := x
	for y != Nil {
		if _, ok := printed[y]; ok {
			count--
			if count < 0 {
				s = append(s, "...") // ellipsis for a circular list
				return strings.Join(s, " ")
			}
		} else {
			printed[y] = true
			count = 4
		}
		s = append(s, str4(y.Car, true, count, printed))
		if cdr, ok := y.Cdr.(*Cell); ok {
			y = cdr
		} else {
			s = append(s, ".")
			s = append(s, str4(y.Cdr, true, count, printed))
			break
		}
	}
	y = x
	for y != Nil {
		delete(printed, y)
		if cdr, ok := y.Cdr.(*Cell); ok {
			y = cdr
		} else {
			break
		}
	}
	return strings.Join(s, " ")
}
