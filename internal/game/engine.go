package game

// Contains game engine code

import (
	"container/list"

	"context"
	"fmt"
	"time"
)

type livingCell struct {
	PosX int `json:"x"`
	PosY int `json:"y"`
}

type livingCellsSet map[livingCell]struct{}

type engine struct {
	generation         int
	deadCells          livingCellsSet
	livingCells        livingCellsSet
	livingCellsHistory *list.List
}

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

func initEngine() *engine {
	return &engine{
		generation:         0,
		deadCells:          make(livingCellsSet, 0),
		livingCells:        make(livingCellsSet, 0),
		livingCellsHistory: list.New(),
	}
}

// TODO: remove renderer after refactor, engine should now know about the renderer
func (e *engine) runGameOfLife(r *renderer, ctx context.Context, pauseChan <-chan bool, millisChan <-chan time.Duration) {
	ticker := time.NewTicker(m * time.Millisecond)
	r.screen.DisableMouse() // disabling mouse before running the game
	defer ticker.Stop()
	defer r.screen.EnableMouse() // enabling mouse before returning
	for {
		select {
		case <-ticker.C:
			e.calcNextGeneration(r)
			r.drawLivingCellsOnGrid()
			r.drawDeadCellsOnGrid()
			// TODO: duplicate code 1
			gameText = fmt.Sprintf("generation: %v, living cells: %v", e.generation, e.livingCells.Len())
			r.drawText(1, gameText)
			r.screen.Show()
		case <-ctx.Done():
			gameText = fmt.Sprintf("stopped after %v generations", e.generation)
			r.drawText(1, gameText)
			r.screen.Show()
			gameIsRuning = false
			return
		case <-pauseChan:
			gameText = fmt.Sprintf("paused after %v generations", e.generation)
			r.drawText(1, gameText)
			r.screen.Show()
			gameIsRuning = false
			return
		case <-millisChan:
			ticker = time.NewTicker(m * time.Millisecond)
		}
	}
}

func (e *engine) calcNextGenDeadCells(lc livingCell) bool {
	count := 0
	for _, d := range directions {
		dx, dy := d[0], d[1]
		if e.livingCells.Contains(livingCell{PosX: lc.PosX + dx, PosY: lc.PosY + dy}) {
			count++
		}
		if count > 3 {
			return false
		}
	}
	return count == 3
	// if count == 3 {
	// 	if lc.PosY > screenOffset && lc.PosY < r.gridHeight-1 && lc.PosX > 0 && lc.PosX < r.gridWidth-1 {
	// 		// TODO: add to renderer.livingCells (after next refactor)
	// 		// e.livingCells.Add(livingCell{PosX: lc.PosX, PosY: lc.PosY})
	// 		// r.screen.Put(lc.PosX, lc.PosY, "@", cs.greenYellow)
	// 	}
	// 	return true
	// }
	// return false
}

// TODO: remove renderer after refactor, engine should now know about the renderer
func (e *engine) calcNextGeneration(r *renderer) {
	if e.livingCellsHistory.Len() > historySize {
		e.livingCellsHistory.Remove(e.livingCellsHistory.Front())
	}
	e.livingCellsHistory.PushBack(e.livingCells)
	livingCellsNextGen := make(livingCellsSet)
	deadCellsNextGen := make(livingCellsSet)
	for lc := range e.livingCells {
		count := 0
		for _, d := range directions {
			dx, dy := d[0], d[1]
			neighborCell := livingCell{PosX: lc.PosX + dx, PosY: lc.PosY + dy}
			if e.livingCells.Contains(neighborCell) {
				count++
			} else {
				ok := e.calcNextGenDeadCells(neighborCell)
				if ok {
					livingCellsNextGen.Add(neighborCell)
				}
			}
		}
		// if count < 2 || count > 3 {
		// 	if lc.PosY > screenOffset && lc.PosY < r.gridHeight-1 && lc.PosX > 0 && lc.PosX < r.gridWidth-1 {
		// 		r.updateCellStyle(lc.PosX, lc.PosY)
		// 	}
		// } else if count == 2 || count == 3 {
		// 	livingCellsNextGen.Add(lc)
		// }
		if count == 2 || count == 3 {
			livingCellsNextGen.Add(lc)
		} else {
			deadCellsNextGen.Add(livingCell{PosX: lc.PosX, PosY: lc.PosY})
			// r.updateCellStyle(lc.PosX, lc.PosY)
		}

	}
	e.livingCells = livingCellsNextGen
	e.deadCells = deadCellsNextGen
	e.generation++
	// gameText = fmt.Sprintf("generation: %v, living cells: %v", e.generation, e.livingCells.Len())
	// r.drawText(1, gameText)
	if e.livingCells.Len() == 0 {
		cancel()
	} // else {
	// r.drawLivingCellsOnGrid()
	// r.drawDeadCellsOnGrid()
	// }
}
