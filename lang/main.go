package lang

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Run executes REPL (Read-Eval-Print Loop).
// It returns false if REPL was ceased by an error.
// It returns true if REPL was finished normally.
func Run(interp *Interp, input io.Reader) bool {
	interactive := (input == nil)
	if interactive {
		input = os.Stdin
	}
	reader := NewReader(input)
	for {
		if interactive {
			os.Stdout.WriteString("> ")
		}
		x, err := reader.Read()
		if err == nil {
			if x == EofToken {
				return true // Finished normally.
			}
			x, err = interp.SafeEval(x, Nil)
			if err == nil {
				if interactive {
					fmt.Println(Str(x))
				}
			}
		}
		if err != nil {
			fmt.Println(err)
			if !interactive {
				return false // Ceased by an error.
			}
		}
	}
}

// Main runs each element of args as a name of Lisp script file.
// It ignores args[0].
// If it does not have args[1] or some element is "-", it begins REPL.
func Main(args []string) int {
	interp := NewInterp()
	ss := strings.NewReader(Prelude)
	if !Run(interp, ss) {
		return 1
	}
	if len(args) < 2 {
		args = []string{args[0], "-"}
	}
	for i, fileName := range args {
		if i == 0 {
			continue
		}
		if fileName == "-" {
			Run(interp, nil)
			fmt.Println("Goodbye")
		} else {
			file, err := os.Open(fileName)
			if err != nil {
				fmt.Println(err)
				return 1
			}
			if !Run(interp, file) {
				return 1
			}
		}
	}
	return 0
}
