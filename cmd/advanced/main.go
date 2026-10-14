package main

import (
	"flag"

	game "go-game-of-life/internal/game"
)

func main() {
	engineType := flag.String("engine", "default", "Engine: default or bit")
	flag.Parse()

	game.Start(*engineType)
}
