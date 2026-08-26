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
	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(cellStyles.def)
	s.EnableMouse()
	s.Clear()

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	loadConfig()
	drawNewGrid(s)
	drawLivingCellsOnGrid(s)

	var (
		lastClickTime time.Time
		lastKeyTime   time.Time
		lastX, lastY  int
		dblClickDelay = 500 * time.Millisecond
		w, h          = s.Size()
		pauseChan     = make(chan bool)
		millisChan    = make(chan time.Duration)
	)

	// Event loop
	for {
		s.Show()           // Updating screen
		ev := <-s.EventQ() // Polling event

		// Processing the events
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h = s.Size()
			if boxOpen {
				drawNewGrid(s)
				drawBox(s, "Choose predefined pattern")
			} else {
				drawNewGrid(s)
				drawLivingCellsOnGrid(s)
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
				livingCells = historyVal.Value.(LivingCellsSet)
				if generation > 0 {
					generation--
				}
				gameText = fmt.Sprintf("history | generation: %v, living cells: %v", generation, livingCells.Len())
				drawNewGrid(s)
				drawLivingCellsOnGrid(s)
			} else if ev.Key() == tcell.KeyRight && !gameIsRuning && !boxOpen {
				calcNextGeneration(s)
				s.Show()
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "b" && !gameIsRuning && len(predefinedLivingCells) > 0 {
				s.DisableMouse()
				boxOpen = true
				historyLivingCells = list.New()
				generation = 0
				drawNewGrid(s)
				livingCells = predefinedLivingCells[predefinedLCIndex%len(predefinedLivingCells)].Copy()
				drawBox(s, "Choose predefined pattern")
				predefinedLCIndex++
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "r" {
				if boxOpen {
					s.EnableMouse()
					drawNewGrid(s)
					drawLivingCellsOnGrid(s)
					drawText(s, 1, "Chosen predefined pattern")
					predefinedLCIndex--
					boxOpen = !boxOpen
				} else if livingCells.Len() == 0 {
					gameText = fmt.Sprintf("not starting, select cells first %v", livingCells.Len())
					drawText(s, 1, gameText)
				} else if !gameIsRuning {
					gameIsRuning = true
					ctx, cancel = context.WithCancel(context.Background())
					go runGameOfLife(s, ctx, pauseChan, millisChan)
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
					x == lastX && y == lastY && y > screenOffset && y < h-1 && x > 0 && x < w-1 { // double-click
					livingCells.Remove(LivingCell{PosX: x, PosY: y})
					gameText = fmt.Sprintf("unselected [%v, %v] - living cells: %v", x, y, livingCells.Len())
					drawText(s, 1, gameText)
					updateCellStyle(s, x, y)
				} else if y > screenOffset && y < h-1 && x > 0 && x < w-1 { // single-click
					livingCells.Add(LivingCell{PosX: x, PosY: y})
					gameText = fmt.Sprintf("selected [%v, %v] - living cells: %v", x, y, livingCells.Len())
					drawText(s, 1, gameText)
					s.Put(x, y, "@", cellStyles.greenYellow)
				}
				lastClickTime = now
				lastX, lastY = x, y
			}
		}
	}
}
