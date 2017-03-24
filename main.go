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

	world = NewWorld()
	if err := world.Init(); err != nil {
		Fatal(err)
	}

	// From now on the screen is initialized so let's handle panics gacefully
	defer handlePanics()

	// Setup initla buffer
	var initialBuffer *Buffer
	if len(os.Args) == 1 {
		initialBuffer = NewBuffer("*scratch*", "")
	} else {
		panic("opening file not implemented")
	}
	world.Buffers = append(world.Buffers, initialBuffer)
	world.Display.WindowTree = NewWindowNode(initialBuffer)
	world.Display.SetCurrentWindow(world.Display.WindowTree)

	world.Run()
}
