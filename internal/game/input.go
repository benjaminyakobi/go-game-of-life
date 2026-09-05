package game

// Contains game input events code

import (
	"fmt"
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

	if ev.Key() == tcell.KeyEscape {
		return true
	}

	handlePause(renderer, engine, state, ev)

	handlePreviousGeneration(renderer, engine, state, ev)

	handleNextGeneration(renderer, engine, state, ev)

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

func handlePause(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev *tcell.EventKey,
) {
	if ev.Key() != tcell.KeyRune ||
		ev.Str() != "p" ||
		state.boxOpen ||
		engine.livingCellsHistory.Len() == 0 {
		return
	}

	if state.running {
		state.running = false

		renderer.screen.EnableMouse()

		gameText = fmt.Sprintf(
			"paused after %v generations",
			engine.generation,
		)

		renderer.drawText(1, gameText)
		renderer.screen.Show()
	}
}

func handlePreviousGeneration(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev *tcell.EventKey,
) {
	if ev.Key() != tcell.KeyLeft ||
		state.running ||
		state.boxOpen ||
		engine.livingCellsHistory.Len() == 0 {
		return
	}

	historyVal := engine.livingCellsHistory.Back()

	engine.livingCellsHistory.Remove(historyVal)

	if engine.generation > 0 {
		engine.generation--
	}

	gameText = fmt.Sprintf(
		"history | generation: %v, living cells: %v",
		engine.generation,
		engine.livingCells.Len(),
	)

	renderer.killLivingCellsOnGrid(engine.livingCells)

	engine.livingCells = historyVal.Value.(cellsSet)

	renderer.drawLivingCellsOnGrid(engine.livingCells)
	renderer.drawText(1, gameText)
	renderer.screen.Show()
}

func handleNextGeneration(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev *tcell.EventKey,
) {
	if ev.Key() != tcell.KeyRight ||
		state.running ||
		state.boxOpen {
		return
	}

	renderer.killLivingCellsOnGrid(engine.livingCells)

	engine.calcNextGeneration()

	renderer.drawLivingCellsOnGrid(engine.livingCells)
	renderer.drawDeadCellsOnGrid(engine.deadCells)

	gameText = fmt.Sprintf(
		"generation: %v, living cells: %v",
		engine.generation,
		engine.livingCells.Len(),
	)

	renderer.drawText(1, gameText)
	renderer.screen.Show()
}
