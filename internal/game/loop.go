package game

// Contains game loop code

import (
	"container/list"
	// "context"
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
		maybePanic := recover()
		renderer.screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	loadConfig()

	renderer.drawNewGrid()
	renderer.drawLivingCellsOnGrid(engine.livingCells) // TODO: explicitly pass engine.livingCells
	renderer.screen.Show()

	var (
		lastClickTime time.Time
		lastKeyTime   time.Time
		lastX, lastY  int
		dblClickDelay               = 500 * time.Millisecond
		running                     = false
		boxOpen                     = false
		m             time.Duration = 500
		ticker                      = time.NewTicker(m * time.Millisecond)
	)

	defer ticker.Stop()

	for {
		select {
		// ============================================================
		// GAME LOOP (originalt from engine.go before the refactor)
		// ============================================================
		case <-ticker.C:
			if !running {
				continue
			}

			// renderer.drawNewGrid()
			renderer.killLivingCellsOnGrid(engine.livingCells)
			engine.calcNextGeneration()
			renderer.drawLivingCellsOnGrid(engine.livingCells) // TODO: explicitly pass engine.livingCells
			renderer.drawDeadCellsOnGrid()                     // TODO: explicitly pass engine.deadCells

			// Stopping when there is no more living cells
			if engine.livingCells.Len() == 0 {
				running = false
				renderer.screen.EnableMouse()
				gameText = fmt.Sprintf(
					"stopped after %v generations",
					engine.generation,
				)
			} else {
				gameText = fmt.Sprintf(
					"generation: %v, living cells: %v",
					engine.generation,
					engine.livingCells.Len(),
				)
			}

			renderer.drawText(1, gameText)
			renderer.screen.Show()

		// ============================================================
		// Tcell EVENT LOOP
		// ============================================================
		case ev := <-renderer.screen.EventQ():
			switch ev := ev.(type) {
			case *tcell.EventResize:
				renderer.drawNewGrid()
				if boxOpen {
					renderer.drawBox("Choose predefined pattern")
				} else {
					renderer.drawLivingCellsOnGrid(engine.livingCells)
				}
				renderer.screen.Show()

			case *tcell.EventKey:
				keyNow := time.Now()

				// Escape
				if ev.Key() == tcell.KeyEscape {
					return
				}

				// Pause
				if ev.Key() == tcell.KeyRune &&
					ev.Str() == "p" &&
					!boxOpen &&
					engine.livingCellsHistory.Len() > 0 {

					if running {
						running = false
						// gameIsRuning = false

						renderer.screen.EnableMouse()

						gameText = fmt.Sprintf(
							"paused after %v generations",
							engine.generation,
						)

						renderer.drawText(1, gameText)
						renderer.screen.Show()
					}
				}

				// Previous generation
				if ev.Key() == tcell.KeyLeft &&
					!running &&
					!boxOpen &&
					engine.livingCellsHistory.Len() > 0 {

					historyVal := engine.livingCellsHistory.Back()

					engine.livingCellsHistory.Remove(historyVal)
					// engine.livingCells = historyVal.Value.(livingCellsSet)

					if engine.generation > 0 {
						engine.generation--
					}

					gameText = fmt.Sprintf(
						"history | generation: %v, living cells: %v",
						engine.generation,
						engine.livingCells.Len(),
					)

					renderer.killLivingCellsOnGrid(engine.livingCells)
					engine.livingCells = historyVal.Value.(livingCellsSet)
					renderer.drawLivingCellsOnGrid(engine.livingCells)
					renderer.drawText(1, gameText)
					renderer.screen.Show()
				}

				// Next generation
				if ev.Key() == tcell.KeyRight &&
					!running &&
					!boxOpen {

					renderer.killLivingCellsOnGrid(engine.livingCells)
					engine.calcNextGeneration()
					renderer.drawLivingCellsOnGrid(engine.livingCells)
					renderer.drawDeadCellsOnGrid()

					gameText = fmt.Sprintf(
						"generation: %v, living cells: %v",
						engine.generation,
						engine.livingCells.Len(),
					)

					renderer.drawText(1, gameText)
					renderer.screen.Show()
				}

				// Choose predefined pattern
				if ev.Key() == tcell.KeyRune &&
					ev.Str() == "b" &&
					!running &&
					len(predefinedLivingCells) > 0 {
					boxOpen = true

					engine.livingCellsHistory = list.New()
					engine.generation = 0

					renderer.screen.DisableMouse()
					renderer.removeBox()
					renderer.killLivingCellsOnGrid(engine.livingCells)
					engine.livingCells = predefinedLivingCells[predefinedLCIndex%len(predefinedLivingCells)].Copy()

					renderer.drawBox("Choose predefined pattern")
					renderer.screen.Show()

					predefinedLCIndex++
				}

				// Run / resume / accept predefined pattern
				if ev.Key() == tcell.KeyRune &&
					ev.Str() == "r" {

					if boxOpen {
						renderer.screen.EnableMouse()

						renderer.removeBox()
						renderer.drawLivingCellsOnGrid(engine.livingCells)
						renderer.drawText(
							1,
							"Chosen predefined pattern",
						)

						predefinedLCIndex--
						boxOpen = false

						renderer.screen.Show()
					} else if engine.livingCells.Len() == 0 {
						gameText = fmt.Sprintf(
							"not starting, select cells first %v",
							engine.livingCells.Len(),
						)

						renderer.drawText(1, gameText)
						renderer.screen.Show()
					} else if !running {
						running = true
						// gameIsRuning = true
						renderer.screen.DisableMouse()
					}
				}

				// Increase speed
				if keyNow.Sub(lastKeyTime) <= dblClickDelay &&
					ev.Key() == tcell.KeyRune &&
					ev.Str() == "=" &&
					running {

					if m > 100 {
						m -= 100

						ticker.Stop()
						ticker = time.NewTicker(
							m * time.Millisecond,
						)
					}
				}

				// Decrease speed
				if keyNow.Sub(lastKeyTime) <= dblClickDelay &&
					ev.Key() == tcell.KeyRune &&
					ev.Str() == "-" &&
					running {

					if m < 1000 {
						m += 100

						ticker.Stop()
						ticker = time.NewTicker(
							m * time.Millisecond,
						)
					}
				}

				// Stop
				if ev.Key() == tcell.KeyRune &&
					ev.Str() == "s" &&
					running {

					running = false
					// gameIsRuning = false

					renderer.screen.EnableMouse()

					gameText = fmt.Sprintf(
						"stopped after %v generations",
						engine.generation,
					)

					renderer.drawText(1, gameText)
					renderer.screen.Show()
				}

				lastKeyTime = keyNow

			case *tcell.EventMouse:
				x, y := ev.Position()

				if ev.Buttons() != tcell.ButtonPrimary {
					continue
				}

				now := time.Now()

				validCell := y > screenOffset &&
					y < renderer.gridHeight-1 &&
					x > 0 &&
					x < renderer.gridWidth-1

				if !validCell || running || boxOpen {
					continue
				}

				// Double click -> remove cell
				if now.Sub(lastClickTime) <= dblClickDelay &&
					x == lastX &&
					y == lastY {

					engine.livingCells.Remove(
						livingCell{
							PosX: x,
							PosY: y,
						},
					)

					gameText = fmt.Sprintf(
						"unselected [%v, %v] - living cells: %v",
						x,
						y,
						engine.livingCells.Len(),
					)

					renderer.drawText(1, gameText)
					renderer.updateCellStyle(x, y)
				} else {
					// Single click -> add cell
					engine.livingCells.Add(
						livingCell{
							PosX: x,
							PosY: y,
						},
					)

					gameText = fmt.Sprintf(
						"selected [%v, %v] - living cells: %v",
						x,
						y,
						engine.livingCells.Len(),
					)

					renderer.drawText(1, gameText)
					renderer.screen.Put(
						x,
						y,
						"@",
						cs.greenYellow,
					) // TODO: this should happend only inside the renderer
				}

				renderer.screen.Show()

				lastClickTime = now
				lastX = x
				lastY = y
			}
		}
	}
}
