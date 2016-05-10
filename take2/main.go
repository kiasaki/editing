package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	Version = "Unknown"
)

var flagVersion = flag.Bool("version", false, "Show editor's version")

var world *World

func main() {
	flag.Parse()
	if *flagVersion {
		fmt.Println("Editor version:", Version)
		os.Exit(0)
	}

	var err error
	world = NewWorld()

	world.Config, err = NewConfig()
	if err != nil {
		Fatal(err)
	}

	world.Display, err = NewDisplay()
	if err != nil {
		Fatal(err)
	}

	err = world.Init()
	if err != nil {
		Fatal(err)
	}

	// From now on the screen is initialized so let's handle panics gacefully
	defer func() {
		err := recover()
		if err != nil {
			Fatal(err.(error))
		}
	}()

	world.Run()
}
