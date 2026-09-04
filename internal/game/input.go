package game

// Contains game input events code

import (
	"github.com/gdamore/tcell/v3"
)

// NOTE: event dispatcher
func handleEvent(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev tcell.Event,
) {
	// switch ev := ev.(type) {
	// case *tcell.EventResize:
	// 	handleResize(renderer, engine, state)
	//
	// case *tcell.EventKey:
	// 	handleKey(renderer, engine, state, ev)
	//
	// case *tcell.EventMouse:
	// 	handleMouse(renderer, engine, state, ev)
	// }
}
