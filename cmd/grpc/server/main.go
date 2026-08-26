package main

import (
	"fmt"
	game "go-game-of-life/internal/game"
)

func main() {
	fmt.Println("grpc server that runs the game stream")
	game.Run()
}
