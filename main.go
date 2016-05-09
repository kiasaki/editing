package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"

	"github.com/kiasaki/editing/config"
	"github.com/kiasaki/editing/display"
	"github.com/kiasaki/editing/text"
)

const (
	Version = "Unknown"
)

var flagVersion = flag.Bool("version", false, "Show editor's version")

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Println("Editor version ", Version)
		os.Exit(0)
	}

	conf := config.ConfigNew()
	_ = text.WorldNew()
	window := display.WindowNew(conf)
	err := window.Init()
	if err != nil {
		Fatal(err)
	}

	// From now on the screen is initialized so let's handle panics gacefully
	defer func() {
		err := recover()
		if err != nil {
			window.End()
			fmt.Print(errors.Wrap(err, 2).ErrorStack())
			Fatal(err.(error))
		}
	}()

	quit := make(chan struct{})

	// Main loop
	go func() {
		for {
			window.Redisplay()

			// Now wait for and handle user event
			ev := window.Screen().PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					window.Refresh()
				}
			case *tcell.EventResize:
				window.Refresh()
			}
		}
	}()

	<-quit

	err = window.End()
	if err != nil {
		Fatal(err)
	}
}

func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
