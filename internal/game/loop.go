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

// TODO: loop only should orchestrate (not knowing about tcell!)
func Run() {
	engine := initEngine()
	renderer, err := initRenderer(engine)
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
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "p" && boxOpen == false && engine.livingCellsHistory.Len() > 0 {
				if gameIsRuning {
					pauseChan <- true
				}
			} else if ev.Key() == tcell.KeyLeft && !gameIsRuning && !boxOpen && engine.livingCellsHistory.Len() > 0 {
				historyVal := engine.livingCellsHistory.Back()
				engine.livingCellsHistory.Remove(historyVal)
				engine.livingCells = historyVal.Value.(livingCellsSet)
				if engine.generation > 0 {
					engine.generation--
				}
				gameText = fmt.Sprintf("history | generation: %v, living cells: %v", engine.generation, engine.livingCells.Len())
				renderer.drawNewGrid()
				renderer.drawLivingCellsOnGrid()
			} else if ev.Key() == tcell.KeyRight && !gameIsRuning && !boxOpen {
				engine.calcNextGeneration()
				renderer.drawLivingCellsOnGrid()
				renderer.drawDeadCellsOnGrid()
				// TODO: duplicate code 1
				gameText = fmt.Sprintf("generation: %v, living cells: %v", engine.generation, engine.livingCells.Len())
				renderer.drawText(1, gameText)
				renderer.screen.Show()
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "b" && !gameIsRuning && len(predefinedLivingCells) > 0 {
				renderer.screen.DisableMouse()
				boxOpen = true
				engine.livingCellsHistory = list.New()
				engine.generation = 0
				renderer.drawNewGrid()
				engine.livingCells = predefinedLivingCells[predefinedLCIndex%len(predefinedLivingCells)].Copy()
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
				} else if engine.livingCells.Len() == 0 {
					gameText = fmt.Sprintf("not starting, select cells first %v", engine.livingCells.Len())
					renderer.drawText(1, gameText)
				} else if !gameIsRuning {
					gameIsRuning = true
					ctx, cancel = context.WithCancel(context.Background())
					go engine.runGameOfLife(renderer, ctx, pauseChan, millisChan)
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
					engine.livingCells.Remove(livingCell{PosX: x, PosY: y})
					gameText = fmt.Sprintf("unselected [%v, %v] - living cells: %v", x, y, engine.livingCells.Len())
					renderer.drawText(1, gameText)
					renderer.updateCellStyle(x, y)
				} else if y > screenOffset && y < renderer.gridHeight-1 && x > 0 && x < renderer.gridWidth-1 { // single-click
					engine.livingCells.Add(livingCell{PosX: x, PosY: y})
					gameText = fmt.Sprintf("selected [%v, %v] - living cells: %v", x, y, engine.livingCells.Len())
					renderer.drawText(1, gameText)
					renderer.screen.Put(x, y, "@", cs.greenYellow)
				}
				lastClickTime = now
				lastX, lastY = x, y
			}
		}
	}
}
