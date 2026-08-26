package game

// Contains game engine code

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v3"
)

type livingCell struct {
	PosX int `json:"x"`
	PosY int `json:"y"`
}

type livingCellsSet map[livingCell]struct{}

func (lcs livingCellsSet) Add(lc livingCell) {
	lcs[lc] = struct{}{}
}

func (lcs livingCellsSet) Remove(lc livingCell) {
	delete(lcs, lc)
}

func (lcs livingCellsSet) Contains(lc livingCell) bool {
	_, ok := lcs[lc]
	return ok
}

func (lcs livingCellsSet) Len() int {
	return len(lcs)
}

func (lcs livingCellsSet) Copy() livingCellsSet {
	if lcs == nil {
		return nil
	}
	lcsCopy := make(livingCellsSet, lcs.Len())
	for cell := range lcs {
		lcsCopy[cell] = struct{}{}
	}
	return lcsCopy
}

func runGameOfLife(s tcell.Screen, ctx context.Context, pauseChan <-chan bool, millisChan <-chan time.Duration) {
	ticker := time.NewTicker(m * time.Millisecond)
	s.DisableMouse() // disabling mouse before running the game
	defer ticker.Stop()
	defer s.EnableMouse() // enabling mouse before returning
	for {
		select {
		case <-ticker.C:
			calcNextGeneration(s)
			s.Show()
		case <-ctx.Done():
			gameText = fmt.Sprintf("stopped after %v generations", generation)
			drawText(s, 1, gameText)
			s.Show()
			gameIsRuning = false
			generation = 0
			return
		case <-pauseChan:
			gameText = fmt.Sprintf("paused after %v generations", generation)
			drawText(s, 1, gameText)
			s.Show()
			gameIsRuning = false
			return
		case <-millisChan:
			ticker = time.NewTicker(m * time.Millisecond)
		}
	}
}

func calcNextGenDeadCells(s tcell.Screen, lc livingCell) bool {
	count := 0
	w, h := s.Size()
	for _, d := range directions {
		dx, dy := d[0], d[1]
		if livingCells.Contains(livingCell{PosX: lc.PosX + dx, PosY: lc.PosY + dy}) {
			count++
		}
		if count > 3 {
			return false
		}
	}
	if count == 3 {
		if lc.PosY > screenOffset && lc.PosY < h-1 && lc.PosX > 0 && lc.PosX < w-1 {
			s.Put(lc.PosX, lc.PosY, "@", cs.greenYellow)
		}
		return true
	}
	return false
}

func calcNextGeneration(s tcell.Screen) {
	w, h := s.Size()
	if historyLivingCells.Len() > historySize {
		historyLivingCells.Remove(historyLivingCells.Front())
	}
	historyLivingCells.PushBack(livingCells)
	livingCellsNextGen := make(livingCellsSet)
	for lc := range livingCells {
		count := 0
		for _, d := range directions {
			dx, dy := d[0], d[1]
			neighborCell := livingCell{PosX: lc.PosX + dx, PosY: lc.PosY + dy}
			if livingCells.Contains(neighborCell) {
				count++
			} else {
				ok := calcNextGenDeadCells(s, neighborCell)
				if ok {
					livingCellsNextGen.Add(neighborCell)
				}
			}
		}
		if count < 2 || count > 3 {
			if lc.PosY > screenOffset && lc.PosY < h-1 && lc.PosX > 0 && lc.PosX < w-1 {
				updateCellStyle(s, lc.PosX, lc.PosY)
			}
		} else if count == 2 || count == 3 {
			livingCellsNextGen.Add(lc)
		}
	}
	livingCells = livingCellsNextGen
	generation++
	gameText = fmt.Sprintf("generation: %v, living cells: %v", generation, livingCells.Len())
	drawText(s, 1, gameText)
	if livingCells.Len() == 0 {
		cancel()
	}
}
