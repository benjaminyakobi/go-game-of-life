package game

import (
	"github.com/gdamore/tcell/v3"
)

type bitRenderer struct {
	screen tcell.Screen
	engine BitEngine
}

func initScreen() (tcell.Screen, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	return screen, nil
}
