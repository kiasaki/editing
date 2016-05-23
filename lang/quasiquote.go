package lang

// QqExpand expands x of any quasi-quote `x into the equivalent S-expression.
func QqExpand(x interface{}) interface{} {
	return qqExpand0(x, 0) // Begin with the nesting level 0.
}

// QqQuote quotes x so that the result evaluates to x.
func QqQuote(x interface{}) interface{} {
	if x == Nil {
		return Nil
	}
	switch x.(type) {
	case *Sym, *Cell:
		return &Cell{QuoteSym, &Cell{x, Nil}}
	default:
		return x
	}
}

func qqExpand0(x interface{}, level int) interface{} {
	if j, ok := x.(*Cell); ok {
		if j == Nil {
			return Nil
		}
		if j.Car == UnquoteSym { // ,a
			if level == 0 {
				return j.CdrCell().Car // ,a => a
			}
		}
		t := qqExpand1(j, level)
		if t.Cdr == Nil {
			if k, ok := t.Car.(*Cell); ok {
				if k.Car == ListSym || k.Car == ConsSym {
					return k
				}
			}
		}
		return &Cell{AppendSym, t}
	} else {
		return QqQuote(x)
	}
}

// qqExpand1 expands x of `x so that the result can be used as an argument of
// append.  Example 1: (,a b) => ((list a 'b))
//          Example 2: (,a ,@(cons 2 3)) => ((cons a (cons 2 3)))
func qqExpand1(x interface{}, level int) *Cell {
	if j, ok := x.(*Cell); ok {
		if j == Nil {
			return &Cell{Nil, Nil}
		}
		switch j.Car {
		case UnquoteSym: // ,a
			if level == 0 {
				return j.CdrCell() // ,a => (a)
			}
			level--
		case QuasiquoteSym: // `a
			level++
		}
		h := qqExpand2(j.Car, level)
		t := qqExpand1(j.Cdr, level) // != Nil
		if t.Car == Nil && t.Cdr == Nil {
			return &Cell{h, Nil}
		} else if hc, ok := h.(*Cell); ok {
			if hc.Car == ListSym {
				if tcar, ok := t.Car.(*Cell); ok {
					if tcar.Car == ListSym {
						hh := qqConcat(hc, tcar.Cdr)
						return &Cell{hh, t.Cdr}
					}
				}
				if hcdr, ok := hc.Cdr.(*Cell); ok {
					hh := qqConsCons(hcdr, t.Car)
					return &Cell{hh, t.Cdr}
				}
			}
		}
		return &Cell{h, t}
	} else {
		return &Cell{QqQuote(x), Nil}
	}
}

// (1 2), (3 4) => (1 2 3 4)
func qqConcat(x *Cell, y interface{}) interface{} {
	if x == Nil {
		return y
	} else {
		return &Cell{x.Car, qqConcat(x.CdrCell(), y)}
	}
}

// (1 2 3), "a" => (cons 1 (cons 2 (cons 3 "a")))
func qqConsCons(x *Cell, y interface{}) interface{} {
	if x == Nil {
		return y
	} else {
		return &Cell{ConsSym, &Cell{x.Car,
			&Cell{qqConsCons(x.CdrCell(), y), Nil}}}
	}
}

// qqExpand2 expands x.car (= y) of `x so that the result can be used as an
// argument of append.
// Examples: ,a => (list a); ,@(foo 1 2) => (foo 1 2); b => (list 'b)
func qqExpand2(y interface{}, level int) interface{} {
	if j, ok := y.(*Cell); ok {
		if j == Nil {
			return &Cell{ListSym, &Cell{Nil, Nil}} // (list nil)
		}
		switch j.Car {
		case UnquoteSym: // ,a
			if level == 0 {
				return &Cell{ListSym, j.Cdr} // ,a => (list a)
			}
			level--
		case UnquoteSplicingSym: // ,@a
			if level == 0 {
				return j.CdrCell().Car // ,@a => a
			}
			level--
		case QuasiquoteSym: // `a
			level++
		}
	}
	return &Cell{ListSym, &Cell{qqExpand0(y, level), Nil}}
}
