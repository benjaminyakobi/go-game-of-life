package game

// Contains game loop code

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v3"
)

func Run() {
	renderer, err := initRenderer()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		renderer.screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	loadConfig()
	renderer.drawNewGrid()
	renderer.drawLivingCellsOnGrid()

	var (
		lastClickTime time.Time
		lastKeyTime   time.Time
		lastX, lastY  int
		dblClickDelay = 500 * time.Millisecond
		pauseChan     = make(chan bool)
		millisChan    = make(chan time.Duration)
	)

	// Event loop
	for {
		renderer.screen.Show()           // Updating screen
		ev := <-renderer.screen.EventQ() // Polling event

		// Processing the events
		switch ev := ev.(type) {
		case *tcell.EventResize:
			if boxOpen {
				renderer.drawNewGrid()
				renderer.drawBox("Choose predefined pattern")
			} else {
				renderer.drawNewGrid()
				renderer.drawLivingCellsOnGrid()
			}
		case *tcell.EventKey:
			keyNow := time.Now()
			if ev.Key() == tcell.KeyEscape {
				return
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "p" && boxOpen == false && historyLivingCells.Len() > 0 {
				if gameIsRuning {
					pauseChan <- true
				}
			} else if ev.Key() == tcell.KeyLeft && !gameIsRuning && !boxOpen && historyLivingCells.Len() > 0 {
				historyVal := historyLivingCells.Back()
				historyLivingCells.Remove(historyVal)
				livingCells = historyVal.Value.(livingCellsSet)
				if generation > 0 {
					generation--
				}
				gameText = fmt.Sprintf("history | generation: %v, living cells: %v", generation, livingCells.Len())
				renderer.drawNewGrid()
				renderer.drawLivingCellsOnGrid()
			} else if ev.Key() == tcell.KeyRight && !gameIsRuning && !boxOpen {
				calcNextGeneration(renderer)
				renderer.screen.Show()
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "b" && !gameIsRuning && len(predefinedLivingCells) > 0 {
				renderer.screen.DisableMouse()
				boxOpen = true
				historyLivingCells = list.New()
				generation = 0
				renderer.drawNewGrid()
				livingCells = predefinedLivingCells[predefinedLCIndex%len(predefinedLivingCells)].Copy()
				renderer.drawBox("Choose predefined pattern")
				predefinedLCIndex++
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "r" {
				if boxOpen {
					renderer.screen.EnableMouse()
					renderer.drawNewGrid()
					renderer.drawLivingCellsOnGrid()
					renderer.drawText(1, "Chosen predefined pattern")
					predefinedLCIndex--
					boxOpen = !boxOpen
				} else if livingCells.Len() == 0 {
					gameText = fmt.Sprintf("not starting, select cells first %v", livingCells.Len())
					renderer.drawText(1, gameText)
				} else if !gameIsRuning {
					gameIsRuning = true
					ctx, cancel = context.WithCancel(context.Background())
					go runGameOfLife(renderer, ctx, pauseChan, millisChan)
				}
			} else if keyNow.Sub(lastKeyTime) <= dblClickDelay && ev.Key() == tcell.KeyRune && ev.Str() == "=" && gameIsRuning {
				if m > 100 {
					m -= 100
					millisChan <- m
				}
			} else if keyNow.Sub(lastKeyTime) <= dblClickDelay && ev.Key() == tcell.KeyRune && ev.Str() == "-" && gameIsRuning {
				if m < 1000 {
					m += 100
					millisChan <- m
				}
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "s" && gameIsRuning {
				cancel()
			}
			lastKeyTime = keyNow
		case *tcell.EventMouse:
			x, y := ev.Position()
			switch ev.Buttons() {
			case tcell.ButtonPrimary:
				now := time.Now()
				if now.Sub(lastClickTime) <= dblClickDelay &&
					x == lastX && y == lastY && y > screenOffset && y < renderer.gridHeight-1 && x > 0 && x < renderer.gridWidth-1 { // double-click
					livingCells.Remove(livingCell{PosX: x, PosY: y})
					gameText = fmt.Sprintf("unselected [%v, %v] - living cells: %v", x, y, livingCells.Len())
					renderer.drawText(1, gameText)
					renderer.updateCellStyle(x, y)
				} else if y > screenOffset && y < renderer.gridHeight-1 && x > 0 && x < renderer.gridWidth-1 { // single-click
					livingCells.Add(livingCell{PosX: x, PosY: y})
					gameText = fmt.Sprintf("selected [%v, %v] - living cells: %v", x, y, livingCells.Len())
					renderer.drawText(1, gameText)
					renderer.screen.Put(x, y, "@", cs.greenYellow)
				}
				lastClickTime = now
				lastX, lastY = x, y
			}
		}
	}
}
