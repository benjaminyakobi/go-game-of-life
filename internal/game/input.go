package game

// Contains game input events code

import (
	"time"

	"github.com/gdamore/tcell/v3"
)

// NOTE: event dispatcher
func handleEvent(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev tcell.Event,
) bool {
	// switch ev := ev.(type) {
	// case *tcell.EventResize:
	// 	handleResize(renderer, engine, state)
	//
	// case *tcell.EventKey:
	// 	return handleKey(renderer, engine, state, ev)
	//
	// case *tcell.EventMouse:
	// 	handleMouse(renderer, engine, state, ev)
	// }

	return false
}

func handleResize(
	renderer *renderer,
	engine *engine,
	state *loopState,
) {
	renderer.drawNewGrid()

	if state.boxOpen {
		renderer.drawBox("Choose predefined pattern")
	} else {
		renderer.drawLivingCellsOnGrid(engine.livingCells)
	}

	renderer.screen.Show()
}

// NOTE: event dispatcher for keyboard events
func handleKey(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev *tcell.EventKey,
) bool {
	keyNow := time.Now()

	// Escape
	if ev.Key() == tcell.KeyEscape {
		// We need to preserve the original return behavior.
		return true
	}

	// Pause
	// handlePause(renderer, engine, state, ev)

	// Previous generation
	// handlePreviousGeneration(renderer, engine, state, ev)

	// Next generation
	// handleNextGeneration(renderer, engine, state, ev)

	// Choose predefined pattern
	// handlePredefinedPattern(renderer, engine, state, ev)

	// Run / resume / accept predefined pattern
	// handleRun(renderer, engine, state, ev)

	// Increase speed
	// handleIncreaseSpeed(renderer, state, ev, keyNow)

	// Decrease speed
	// handleDecreaseSpeed(renderer, state, ev, keyNow)

	// Stop
	// handleStop(renderer, engine, state, ev)

	state.lastKeyTime = keyNow

	return false
}
