package game

// Contains game input events code

import (
	"fmt"
	"time"
	// "github.com/gdamore/tcell/v3" // TODO: remove dependency
)

// // NOTE: event dispatcher
// func handleEvent(
// 	renderer *renderer,
// 	state *loopState,
// 	ev tcell.Event,
// ) bool { // TODO: move to renderer.go as a method not a function
// 	switch ev := ev.(type) {
// 	case *tcell.EventResize:
// 		handleResize(renderer, state)
//
// 	case *tcell.EventKey:
// 		return handleKey(renderer, state, ev)
//
// 	case *tcell.EventMouse:
// 		handleMouse(renderer, state, ev)
// 	}
//
// 	return false
// }

func handleResize(
	renderer *renderer,
	state *loopState,
) {
	renderer.drawNewGrid()

	if state.boxOpen {
		renderer.drawBox(renderer.engine.PatternName())
	} else {
		cs := renderer.centerCells(
			renderer.engine.GetLivingCells(),
			0, 0, renderer.gridWidth, renderer.gridHeight)
		renderer.engine.SetLivingCells(cs)
		renderer.drawLivingCellsOnGrid(cs)

		renderer.centerCellsHistory(renderer.engine.LivingCellsHistory())
	}

	renderer.screen.Show()
}

// NOTE: event dispatcher for keyboard events
// func handleKey(
// 	renderer *renderer,
// 	state *loopState,
// 	ev *tcell.EventKey,
// ) bool {
// 	// keyNow := time.Now()
// 	//
// 	// if ev.Key() == tcell.KeyEscape {
// 	// 	return true
// 	// }
//
// 	// handlePause(renderer, state, ev)
//
// 	// handlePreviousGeneration(renderer, state, ev)
//
// 	// handleNextGeneration(renderer, state, ev)
//
// 	// handleNextPredefinedPattern(renderer, state, ev)
//
// 	// handlePreviousPredefinedPattern(renderer, state, ev)
//
// 	// Run / resume / accept predefined pattern
// 	// handleRun(renderer, state, ev)
//
// 	// handleIncreaseSpeed(state, ev, keyNow)
// 	//
// 	// handleDecreaseSpeed(state, ev, keyNow)
// 	//
// 	// handleStop(renderer, state, ev)
// 	//
// 	// state.lastKeyTime = keyNow
//
// 	return false
// }

func handlePause(
	renderer *renderer,
	state *loopState,
	// ev *tcell.EventKey,
) {
	// if ev.Key() != tcell.KeyRune ||
	// 	ev.Str() != "p" ||
	// 	state.boxOpen {
	// 	return
	// }
	if state.boxOpen || !state.running {
		return
	}

	// if state.running {
	state.running = false

	renderer.screen.EnableMouse()

	renderer.drawText(1, fmt.Sprintf(
		"paused after %v generations",
		renderer.engine.GetGeneration(),
	))
	renderer.screen.Show()
	// }
}

func handlePreviousGeneration(
	renderer *renderer,
	state *loopState,
	// ev *tcell.EventKey,
) {
	// if ev.Key() != tcell.KeyLeft ||
	// 	state.running ||
	// 	state.boxOpen {
	// 	return
	// }
	if state.running || state.boxOpen {
		return
	}

	historyVal := renderer.engine.PreviousGeneration()
	if historyVal == nil {
		return
	}

	if renderer.engine.GetGeneration() > 0 {
		renderer.engine.DecrementGeneration()
	}

	renderer.drawText(1, fmt.Sprintf(
		"history | generation: %v, living cells: %v",
		renderer.engine.GetGeneration(),
		renderer.engine.GetLivingCells().Len(),
	))

	renderer.drawDeadCellsOnGrid(renderer.engine.GetLivingCells())

	renderer.engine.SetLivingCells(historyVal)

	renderer.drawLivingCellsOnGrid(renderer.engine.GetLivingCells())
	renderer.screen.Show()
}

func handleNextGeneration(
	renderer *renderer,
	state *loopState,
	// ev *tcell.EventKey,
) {
	// if ev.Key() != tcell.KeyRight ||
	// 	state.running ||
	// 	state.boxOpen {
	// 	return
	// }
	if state.running || state.boxOpen {
		return
	}

	renderer.engine.NextGeneration()

	renderer.drawDeadCellsOnGrid(renderer.engine.DeadCells())
	renderer.drawLivingCellsOnGrid(renderer.engine.GetLivingCells())

	renderer.drawText(1, fmt.Sprintf(
		"generation: %v, living cells: %v",
		renderer.engine.GetGeneration(),
		renderer.engine.GetLivingCells().Len(),
	))
	renderer.screen.Show()
}

func handleNextPredefinedPattern(
	renderer *renderer,
	state *loopState,
	// ev *tcell.EventKey,
) {
	// if ev.Key() != tcell.KeyRune ||
	// 	ev.Str() != "b" ||
	// 	state.running ||
	// 	renderer.engine.Patterns().Len() == 0 {
	// 	return
	// }
	if state.running || renderer.engine.Patterns().Len() == 0 {
		return
	}

	state.boxOpen = true

	lch := renderer.engine.LivingCellsHistory()
	lch.Clear()
	// renderer.engine.LivingCellsHistory().Clear()
	renderer.engine.ResetGeneration()

	renderer.screen.DisableMouse()
	renderer.removeBox()

	renderer.drawDeadCellsOnGrid(renderer.engine.GetLivingCells())

	state.boxIndex++
	if state.boxIndex >= renderer.engine.Patterns().Len() {
		state.boxIndex = 0
	}
	patternName, livingCells := renderer.engine.Patterns().Get(state.boxIndex)
	renderer.engine.SetLivingCells(livingCells)
	renderer.engine.SetPatternName(patternName)

	renderer.drawBox(renderer.engine.PatternName())
	renderer.screen.Show()

}

func handlePreviousPredefinedPattern(
	renderer *renderer,
	state *loopState,
	// ev *tcell.EventKey,
) {
	// if ev.Key() != tcell.KeyRune ||
	// 	ev.Str() != "B" ||
	// 	state.running ||
	// 	renderer.engine.Patterns().Len() == 0 {
	// 	return
	// }
	if state.running || renderer.engine.Patterns().Len() == 0 {
		return
	}

	state.boxOpen = true

	lch := renderer.engine.LivingCellsHistory()
	lch.Clear()
	// renderer.engine.LivingCellsHistory().Clear()
	renderer.engine.ResetGeneration()

	renderer.screen.DisableMouse()
	renderer.removeBox()

	renderer.drawDeadCellsOnGrid(renderer.engine.GetLivingCells())

	state.boxIndex--
	if state.boxIndex < 0 {
		state.boxIndex = renderer.engine.Patterns().Len() - 1
	}
	patternName, livingCells := renderer.engine.Patterns().Get(state.boxIndex)
	renderer.engine.SetLivingCells(livingCells)
	renderer.engine.SetPatternName(patternName)

	renderer.drawBox(renderer.engine.PatternName())
	renderer.screen.Show()

}

func handleRun(
	renderer *renderer,
	state *loopState,
	// ev *tcell.EventKey,
) {
	// if ev.Key() != tcell.KeyRune || ev.Str() != "r" {
	// 	return
	// }

	if state.boxOpen {
		renderer.screen.EnableMouse()

		renderer.removeBox()
		renderer.drawLivingCellsOnGrid(renderer.engine.GetLivingCells())

		state.boxOpen = false

		renderer.screen.Show()

		return
	}

	if renderer.engine.GetLivingCells().Len() == 0 {
		renderer.drawText(1, fmt.Sprintf(
			"not starting, select cells first %v",
			renderer.engine.GetLivingCells().Len(),
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
	// ev *tcell.EventKey,
	// keyNow time.Time,
) {
	// if keyNow.Sub(state.lastKeyTime) > state.dblClickDelay ||
	// 	ev.Key() != tcell.KeyRune ||
	// 	ev.Str() != "=" ||
	// 	!state.running {
	// 	return
	// }

	if state.interval > 100*time.Millisecond {
		state.interval -= 100 * time.Millisecond
		state.ticker.Reset(state.interval)
	}
}

func handleDecreaseSpeed(
	state *loopState,
	// ev *tcell.EventKey,
	// keyNow time.Time,
) {
	// if keyNow.Sub(state.lastKeyTime) > state.dblClickDelay ||
	// 	ev.Key() != tcell.KeyRune ||
	// 	ev.Str() != "-" ||
	// 	!state.running {
	// 	return
	// }

	if state.interval < 1000*time.Millisecond {
		state.interval += 100 * time.Millisecond
		state.ticker.Reset(state.interval)
	}
}

func handleStop(
	renderer *renderer,
	state *loopState,
	// ev *tcell.EventKey,
) {
	// if ev.Key() != tcell.KeyRune ||
	// 	ev.Str() != "s" ||
	// 	!state.running {
	// 	return
	// }
	if !state.running {
		return
	}

	state.running = false

	renderer.screen.EnableMouse()

	renderer.drawText(1, fmt.Sprintf(
		"stopped after %v generations",
		renderer.engine.GetGeneration(),
	))
	renderer.screen.Show()
}

func handleMouse(
	renderer *renderer,
	state *loopState,
	x, y int,
	// ev *tcell.EventMouse,
) {
	// x, y := ev.Position()
	//
	// if ev.Buttons() != tcell.ButtonPrimary {
	// 	return
	// }

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
		renderer.engine.GetLivingCells().Remove(c)
		renderer.drawSingleDeadCellOnGrid(c)

		renderer.drawText(1, fmt.Sprintf(
			"unselected [%v, %v] - living cells: %v",
			c.PosX,
			c.PosY,
			renderer.engine.GetLivingCells().Len(),
		))
	} else {
		renderer.engine.GetLivingCells().Add(c)
		renderer.drawSingleLivingCellOnGrid(c)

		renderer.drawText(1, fmt.Sprintf(
			"selected [%v, %v] - living cells: %v",
			c.PosX,
			c.PosY,
			renderer.engine.GetLivingCells().Len(),
		))
	}

	renderer.screen.Show()
	state.lastClickTime = now
	state.lastX = x
	state.lastY = y
}
