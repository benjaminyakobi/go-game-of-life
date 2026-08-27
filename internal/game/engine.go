package game

// Contains game engine code

import (
	"context"
	"fmt"
	"time"
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

func runGameOfLife(r *renderer, ctx context.Context, pauseChan <-chan bool, millisChan <-chan time.Duration) {
	ticker := time.NewTicker(m * time.Millisecond)
	r.screen.DisableMouse() // disabling mouse before running the game
	defer ticker.Stop()
	defer r.screen.EnableMouse() // enabling mouse before returning
	for {
		select {
		case <-ticker.C:
			calcNextGeneration(r)
			r.screen.Show()
		case <-ctx.Done():
			gameText = fmt.Sprintf("stopped after %v generations", generation)
			r.drawText(1, gameText)
			r.screen.Show()
			gameIsRuning = false
			generation = 0
			return
		case <-pauseChan:
			gameText = fmt.Sprintf("paused after %v generations", generation)
			r.drawText(1, gameText)
			r.screen.Show()
			gameIsRuning = false
			return
		case <-millisChan:
			ticker = time.NewTicker(m * time.Millisecond)
		}
	}
}

func calcNextGenDeadCells(r *renderer, lc livingCell) bool {
	count := 0
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
		if lc.PosY > screenOffset && lc.PosY < r.gridHeight-1 && lc.PosX > 0 && lc.PosX < r.gridWidth-1 {
			r.screen.Put(lc.PosX, lc.PosY, "@", cs.greenYellow)
		}
		return true
	}
	return false
}

func calcNextGeneration(r *renderer) {
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
				ok := calcNextGenDeadCells(r, neighborCell)
				if ok {
					livingCellsNextGen.Add(neighborCell)
				}
			}
		}
		if count < 2 || count > 3 {
			if lc.PosY > screenOffset && lc.PosY < r.gridHeight-1 && lc.PosX > 0 && lc.PosX < r.gridWidth-1 {
				r.updateCellStyle(lc.PosX, lc.PosY)
			}
		} else if count == 2 || count == 3 {
			livingCellsNextGen.Add(lc)
		}
	}
	livingCells = livingCellsNextGen
	generation++
	gameText = fmt.Sprintf("generation: %v, living cells: %v", generation, livingCells.Len())
	r.drawText(1, gameText)
	if livingCells.Len() == 0 {
		cancel()
	}
}
