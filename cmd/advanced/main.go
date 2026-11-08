package main

import (
	"flag"
	"fmt"
	"os"

	game "go-game-of-life/internal/game"
	onedgame "go-game-of-life/internal/onedgame"
)

func main() {
	engineType := flag.String("engine", "default", "Engine: default or 1d")
	flag.Parse()

	switch *engineType {
	case "default":
		game.Start()
	case "1d":
		onedgame.Start()
	default:
		fmt.Fprintln(os.Stderr, "unknown engine:", *engineType)
		os.Exit(1)
	}
}
