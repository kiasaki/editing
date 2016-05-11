package lang

import (
	"fmt"
	"strings"
)

type SValue interface {
	String() string
	Type() string
}

type SInt int64

func (_ SInt) Type() string {
	return "integer"
}

func (this SInt) String() string {
	return fmt.Sprintf("%d", int64(this))
}

type SNum float64

func (_ SNum) Type() string {
	return "number"
}

func (this SNum) String() string {
	return fmt.Sprintf("%f", float64(this))
}

type SChar rune

func (_ SChar) Type() string {
	return "char"
}

func (this SChar) String() string {
	return "\\" + string(this)
}

type SStr string

func (_ SStr) Type() string {
	return "string"
}

func (this SStr) String() string {
	return "\"" + string(this) + "\""
}

type SList []SValue

func (_ SList) Type() string {
	return "list"
}

func (this SList) String() string {
	listString := ""
	for _, element := range this {
		listString = listString + " " + element.String()
	}
	return "(" + strings.Trim(listString, " ") + ")"
}
