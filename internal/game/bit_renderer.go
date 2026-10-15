package game

import (
	"github.com/gdamore/tcell/v3"
)

type bitRenderer struct {
	screen tcell.Screen
	engine BitEngine
}
