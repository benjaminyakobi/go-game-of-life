package game

// Contains game input events code

import (
	"container/list"
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

	handlePredefinedPattern(renderer, engine, state, ev)

	// Run / resume / accept predefined pattern
	handleRun(renderer, engine, state, ev)

	handleIncreaseSpeed(state, ev, keyNow)

	handleDecreaseSpeed(state, ev, keyNow)

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

func handlePredefinedPattern(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev *tcell.EventKey,
) {
	if ev.Key() != tcell.KeyRune ||
		ev.Str() != "b" ||
		state.running ||
		len(predefinedLivingCells) == 0 {
		return
	}

	state.boxOpen = true

	engine.livingCellsHistory = list.New()
	engine.generation = 0

	renderer.screen.DisableMouse()
	renderer.removeBox()
	renderer.killLivingCellsOnGrid(engine.livingCells)

	engine.livingCells =
		predefinedLivingCells[predefinedLCIndex%len(predefinedLivingCells)].Copy()

	renderer.drawBox("Choose predefined pattern")
	renderer.screen.Show()

	predefinedLCIndex++
}

func handleRun(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev *tcell.EventKey,
) {
	if ev.Key() != tcell.KeyRune || ev.Str() != "r" {
		return
	}

	if state.boxOpen {
		renderer.screen.EnableMouse()

		renderer.removeBox()
		renderer.drawLivingCellsOnGrid(engine.livingCells)
		renderer.drawText(
			1,
			"Chosen predefined pattern",
		)

		predefinedLCIndex--
		state.boxOpen = false

		renderer.screen.Show()

		return
	}

	if engine.livingCells.Len() == 0 {
		gameText = fmt.Sprintf(
			"not starting, select cells first %v",
			engine.livingCells.Len(),
		)

		renderer.drawText(1, gameText)
		renderer.screen.Show()

		return
	}

	if !state.running {
		state.running = true
		renderer.screen.DisableMouse()
	}
}

func handleIncreaseSpeed(
	state *loopState,
	ev *tcell.EventKey,
	keyNow time.Time,
) {
	if keyNow.Sub(state.lastKeyTime) > state.dblClickDelay ||
		ev.Key() != tcell.KeyRune ||
		ev.Str() != "=" ||
		!state.running {
		return
	}

	if state.interval > 100 {
		state.interval -= 100

		state.ticker.Stop()
		state.ticker = time.NewTicker(
			state.interval * time.Millisecond,
		)
	}
}

func handleDecreaseSpeed(
	state *loopState,
	ev *tcell.EventKey,
	keyNow time.Time,
) {
	if keyNow.Sub(state.lastKeyTime) > state.dblClickDelay ||
		ev.Key() != tcell.KeyRune ||
		ev.Str() != "-" ||
		!state.running {
		return
	}

	if state.interval < 1000 {
		state.interval += 100

		state.ticker.Stop()
		state.ticker = time.NewTicker(
			state.interval * time.Millisecond,
		)
	}
}
