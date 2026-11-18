package main

import (
	"flag"
	"fmt"
	"os"

	coo "go-game-of-life/internal/coordinates"
	lin "go-game-of-life/internal/linear"
)

func main() {
	engineType := flag.String("engine", "default", "Engine: default or 1d")
	flag.Parse()

	switch *engineType {
	case "default":
		coo.Start()
	case "1d":
		lin.Start()
	default:
		fmt.Fprintln(os.Stderr, "unknown engine:", *engineType)
		os.Exit(1)
	}
}
