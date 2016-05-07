package main

import (
	"fmt"

	"github.com/kiasaki/editing/lang"
)

func main() {
	fmt.Println(lang.SList([]lang.SValue{
		lang.SStr("ahh!"), lang.SNum(1), lang.SInt(9),
		lang.SList([]lang.SValue{
			lang.SChar('H'),
			lang.SChar('e'),
			lang.SChar('l'),
			lang.SChar('l'),
			lang.SChar('o'),
		}),
	}))
}
