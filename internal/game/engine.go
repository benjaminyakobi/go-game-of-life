package game

// Contains game engine code

import (
	"container/list"
	// "context"
	// "fmt"
	// "time"
)

type cell struct {
	PosX int `json:"x"`
	PosY int `json:"y"`
}

type cellsSet map[cell]struct{}

type engine struct {
	generation         int
	deadCells          cellsSet
	livingCells        cellsSet
	livingCellsHistory *list.List // TODO: convert to slice
}

func (cs cellsSet) Add(c cell) {
	cs[c] = struct{}{}
}

func (cs cellsSet) Remove(c cell) {
	delete(cs, c)
}

func (cs cellsSet) Contains(c cell) bool {
	_, ok := cs[c]
	return ok
}

func (cs cellsSet) Len() int {
	return len(cs)
}

func (cs cellsSet) Copy() cellsSet {
	if cs == nil {
		return nil
	}
	csCopy := make(cellsSet, cs.Len())
	for c := range cs {
		csCopy[c] = struct{}{}
	}
	return csCopy
}

func initEngine() *engine {
	return &engine{
		generation:         0,
		deadCells:          make(cellsSet, 0),
		livingCells:        make(cellsSet, 0),
		livingCellsHistory: list.New(),
	}
}

// TODO: remove renderer after refactor, engine should now know about the renderer
// func (e *engine) runGameOfLife(r *renderer, ctx context.Context, pauseChan <-chan bool, millisChan <-chan time.Duration) {
// 	ticker := time.NewTicker(m * time.Millisecond)
// 	r.screen.DisableMouse() // disabling mouse before running the game
// 	defer ticker.Stop()
// 	defer r.screen.EnableMouse() // enabling mouse before returning
// 	for {
// 		select {
// 		case <-ticker.C:
// 			e.calcNextGeneration()
// 			r.drawLivingCellsOnGrid()
// 			r.drawDeadCellsOnGrid()
// 			// TODO: duplicate code 1
// 			gameText = fmt.Sprintf("generation: %v, living cells: %v", e.generation, e.livingCells.Len())
// 			r.drawText(1, gameText)
// 			r.screen.Show()
// 		case <-ctx.Done():
// 			gameText = fmt.Sprintf("stopped after %v generations", e.generation)
// 			r.drawText(1, gameText)
// 			r.screen.Show()
// 			gameIsRuning = false
// 			return
// 		case <-pauseChan:
// 			gameText = fmt.Sprintf("paused after %v generations", e.generation)
// 			r.drawText(1, gameText)
// 			r.screen.Show()
// 			gameIsRuning = false
// 			return
// 		case <-millisChan:
// 			ticker.Stop()
// 			ticker = time.NewTicker(m * time.Millisecond)
// 		}
// 	}
// }

func (e *engine) calcNextGenDeadCells(lc cell) bool {
	count := 0
	for _, d := range directions {
		dx, dy := d[0], d[1]
		if e.livingCells.Contains(cell{PosX: lc.PosX + dx, PosY: lc.PosY + dy}) {
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

func (e *engine) calcNextGeneration() {
	if e.livingCellsHistory.Len() > historySize {
		e.livingCellsHistory.Remove(e.livingCellsHistory.Front())
	}
	e.livingCellsHistory.PushBack(e.livingCells)
	livingCellsNextGen := make(cellsSet)
	deadCellsNextGen := make(cellsSet)
	for lc := range e.livingCells {
		count := 0
		for _, d := range directions {
			dx, dy := d[0], d[1]
			neighborCell := cell{PosX: lc.PosX + dx, PosY: lc.PosY + dy}
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
			deadCellsNextGen.Add(cell{PosX: lc.PosX, PosY: lc.PosY})
			// r.updateCellStyle(lc.PosX, lc.PosY)
		}

	}
	e.livingCells = livingCellsNextGen
	e.deadCells = deadCellsNextGen
	e.generation++
	// gameText = fmt.Sprintf("generation: %v, living cells: %v", e.generation, e.livingCells.Len())
	// r.drawText(1, gameText)
	// if e.livingCells.Len() == 0 {
	// 	cancel()
	//	} // else {
	// r.drawLivingCellsOnGrid()
	// r.drawDeadCellsOnGrid()
	// }
}
