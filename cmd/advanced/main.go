package main

import (
	"flag"
	"fmt"
	"os"

	coo "go-game-of-life/internal/coordinates"
	lin "go-game-of-life/internal/linear"
)

func main() {
	engineType := flag.String("engine", "default", "Engine: default or linear")
	flag.Parse()

	switch *engineType {
	case "default":
		coo.Start()
	case "linear":
		lin.Start()
	default:
		fmt.Fprintln(os.Stderr, "unknown engine:", *engineType)
		os.Exit(1)
	}
}
