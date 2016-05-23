package lang

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

// Reader represents a reader of Lisp expressions.
type Reader struct {
	scanner *bufio.Scanner
	token   interface{} // the current token
	tokens  []string    // tokens read from the current line
	index   int         // the next index of tokens
	line    string      // the current line
	lineNo  int         // the current line number
	erred   bool        // a flag if an error has happened
}

// NewReader constructs a reader which will read Lisp expressions from r.
func NewReader(r io.Reader) *Reader {
	scanner := bufio.NewScanner(r)
	return &Reader{scanner, nil, nil, 0, "", 0, false}
}

// Read reads a Lisp expression and returns the expression and nil.
// If the input runs out, it returns EofToken and nil.
// If an error happens, it returns Nil and the error.
func (rr *Reader) Read() (result interface{}, err interface{}) {
	defer func() {
		if e := recover(); e != nil {
			result, err = Nil, e
		}
	}()
	rr.readToken()
	return rr.parseExpression(), nil
}

func (rr *Reader) newSynatxError(msg string, arg interface{}) *EvalError {
	rr.erred = true
	s := fmt.Sprintf("syntax error: %s -- %d: %s",
		fmt.Sprintf(msg, arg), rr.lineNo, rr.line)
	return &EvalError{s, nil}
}

func (rr *Reader) parseExpression() interface{} {
	switch rr.token {
	case LeftParenSym: // (a b c)
		rr.readToken()
		return rr.parseListBody()
	case SingleQuoteSym: // 'a => (quote a)
		rr.readToken()
		return &Cell{QuoteSym, &Cell{rr.parseExpression(), Nil}}
	case BackQuoteSym: // `a => (quasiquote a)
		rr.readToken()
		return &Cell{QuasiquoteSym, &Cell{rr.parseExpression(), Nil}}
	case CommaSym: // ,a => (unquote a)
		rr.readToken()
		return &Cell{UnquoteSym, &Cell{rr.parseExpression(), Nil}}
	case CommaAtSym: // ,@a => (unquote-splicing a)
		rr.readToken()
		return &Cell{UnquoteSplicingSym, &Cell{rr.parseExpression(), Nil}}
	case DotSym, RightParenSym:
		panic(rr.newSynatxError("unexpected \"%v\"", rr.token))
	default:
		return rr.token
	}
}

func (rr *Reader) parseListBody() *Cell {
	if rr.token == EofToken {
		panic(rr.newSynatxError("unexpected EOF%s", ""))
	} else if rr.token == RightParenSym {
		return Nil
	} else {
		e1 := rr.parseExpression()
		rr.readToken()
		var e2 interface{}
		if rr.token == DotSym { // (a . b)
			rr.readToken()
			e2 = rr.parseExpression()
			rr.readToken()
			if rr.token != RightParenSym {
				panic(rr.newSynatxError("\")\" expected: %v", rr.token))
			}
		} else {
			e2 = rr.parseListBody()
		}
		return &Cell{e1, e2}
	}
}

// readToken reads the next token and set it to rr.token.
func (rr *Reader) readToken() {
	// Read the next line if the line ends or an error happened last time.
	for len(rr.tokens) <= rr.index || rr.erred {
		rr.erred = false
		if rr.scanner.Scan() {
			rr.line = rr.scanner.Text()
			rr.lineNo++
		} else {
			if err := rr.scanner.Err(); err != nil {
				panic(err)
			}
			rr.token = EofToken
			return
		}
		mm := tokenPat.FindAllStringSubmatch(rr.line, -1)
		tt := make([]string, 0, len(mm)*3/5) // Estimate 40% will be spaces.
		for _, m := range mm {
			if m[1] != "" {
				tt = append(tt, m[1])
			}
		}
		rr.tokens = tt
		rr.index = 0
	}
	// Read the next token.
	s := rr.tokens[rr.index]
	rr.index++
	if s[0] == '"' {
		n := len(s) - 1
		if n < 1 || s[n] != '"' {
			panic(rr.newSynatxError("bad string: '%s'", s))
		}
		s = s[1:n]
		s = escapePat.ReplaceAllStringFunc(s, func(t string) string {
			r, ok := escapes[t] // r, err := strconv.Unquote("'" + t + "'")
			if !ok {
				r = t // Leave any invalid escape sequence as it is.
			}
			return r
		})
		rr.token = s
		return
	}
	f, err := strconv.ParseFloat(s, 64) // Try to read s as a float64.
	if err == nil {
		rr.token = f
		return
	}
	if s == "nil" {
		rr.token = Nil
		return
	} else if s == "t" {
		rr.token = true
		return
	}
	rr.token = NewSym(s)
	return
}

// tokenPat is a regular expression to split a line to Lisp tokens.
var tokenPat = regexp.MustCompile(`\s+|;.*$|("(\\.?|.)*?"|,@?|[^()'` +
	"`" + `~"; ]+|.)`)

// escapePat is a reg. expression to take an escape sequence out of a string.
var escapePat = regexp.MustCompile(`\\(.)`)

// escapes is a mapping from an escape sequence to its string value.
var escapes = map[string]string{
	`\\`: `\`,
	`\"`: `"`,
	`\n`: "\n", `\r`: "\r", `\f`: "\f", `\b`: "\b", `\t`: "\t", `\v`: "\v",
}
