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
	switch ev := ev.(type) {
	case *tcell.EventResize:
		handleResize(renderer, engine, state)

	case *tcell.EventKey:
		return handleKey(renderer, engine, state, ev)

	case *tcell.EventMouse:
		handleMouse(renderer, engine, state, ev)
	}

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
		renderer.engine.livingCells = renderer.centerCells(
			renderer.engine.livingCells,
			0, 0, renderer.gridWidth, renderer.gridHeight)
		renderer.drawLivingCellsOnGrid(engine.livingCells)
		// TODO: improve this part, currently seems to work
		centeredHistory := make([]cellsSet, 0)
		for _, cs := range engine.livingCellsHistory {
			centeredCS := renderer.centerCells(cs, 0, 0, renderer.gridWidth, renderer.gridHeight)
			centeredHistory = append(centeredHistory, centeredCS)
		}
		renderer.engine.livingCellsHistory = centeredHistory
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

	handleStop(renderer, engine, state, ev)

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
		len(engine.livingCellsHistory) == 0 {
		return
	}

	if state.running {
		state.running = false

		renderer.screen.EnableMouse()

		renderer.drawText(1, fmt.Sprintf(
			"paused after %v generations",
			engine.generation,
		))
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
		len(engine.livingCellsHistory) == 0 {
		return
	}

	lastHistoryIndex := len(engine.livingCellsHistory) - 1
	historyVal := engine.livingCellsHistory[lastHistoryIndex]
	engine.livingCellsHistory = engine.livingCellsHistory[:lastHistoryIndex]

	if engine.generation > 0 {
		engine.generation--
	}

	renderer.drawText(1, fmt.Sprintf(
		"history | generation: %v, living cells: %v",
		engine.generation,
		engine.livingCells.Len(),
	))
	renderer.killLivingCellsOnGrid(engine.livingCells)

	engine.livingCells = historyVal

	renderer.drawLivingCellsOnGrid(engine.livingCells)
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

	renderer.drawText(1, fmt.Sprintf(
		"generation: %v, living cells: %v",
		engine.generation,
		engine.livingCells.Len(),
	))
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
		len(engine.patterns) == 0 {
		return
	}

	state.boxOpen = true

	engine.livingCellsHistory = make([]cellsSet, 0)
	engine.generation = 0

	renderer.screen.DisableMouse()
	renderer.removeBox()
	renderer.killLivingCellsOnGrid(engine.livingCells)

	engine.livingCells =
		engine.patterns[state.boxIndex%len(engine.patterns)].Copy()

	renderer.drawBox("Choose predefined pattern")
	renderer.screen.Show()

	state.boxIndex++
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

		state.boxIndex--
		state.boxOpen = false

		renderer.screen.Show()

		return
	}

	if engine.livingCells.Len() == 0 {
		renderer.drawText(1, fmt.Sprintf(
			"not starting, select cells first %v",
			engine.livingCells.Len(),
		))
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

	if state.interval > 100*time.Millisecond {
		state.interval -= 100 * time.Millisecond
		state.ticker.Reset(state.interval)
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

	if state.interval < 1000*time.Millisecond {
		state.interval += 100 * time.Millisecond
		state.ticker.Reset(state.interval)
	}
}

func handleStop(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev *tcell.EventKey,
) {
	if ev.Key() != tcell.KeyRune ||
		ev.Str() != "s" ||
		!state.running {
		return
	}

	state.running = false

	renderer.screen.EnableMouse()

	renderer.drawText(1, fmt.Sprintf(
		"stopped after %v generations",
		engine.generation,
	))
	renderer.screen.Show()
}

func handleMouse(
	renderer *renderer,
	engine *engine,
	state *loopState,
	ev *tcell.EventMouse,
) {
	x, y := ev.Position()

	if ev.Buttons() != tcell.ButtonPrimary {
		return
	}

	now := time.Now()

	validCell := y > renderer.gridOffset &&
		y < renderer.gridHeight-1 &&
		x > 0 &&
		x < renderer.gridWidth-1

	if !validCell || state.running || state.boxOpen {
		return
	}

	c := cell{PosX: x, PosY: y}

	if now.Sub(state.lastClickTime) <= state.dblClickDelay &&
		c.PosX == state.lastX &&
		c.PosY == state.lastY {
		engine.livingCells.Remove(c)
		renderer.drawSingleDeadCellOnGrid(c)

		renderer.drawText(1, fmt.Sprintf(
			"unselected [%v, %v] - living cells: %v",
			c.PosX,
			c.PosY,
			engine.livingCells.Len(),
		))
	} else {
		engine.livingCells.Add(c)
		renderer.drawSingleLivingCellOnGrid(c)

		renderer.drawText(1, fmt.Sprintf(
			"selected [%v, %v] - living cells: %v",
			c.PosX,
			c.PosY,
			engine.livingCells.Len(),
		))
	}

	renderer.screen.Show()
	state.lastClickTime = now
	state.lastX = x
	state.lastY = y
}
