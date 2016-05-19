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

	world.Run()
}
