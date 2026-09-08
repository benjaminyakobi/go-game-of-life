package game

// Contains game engine code

import (
	"container/list"
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
		if count == 2 || count == 3 {
			livingCellsNextGen.Add(lc)
		} else {
			deadCellsNextGen.Add(cell{PosX: lc.PosX, PosY: lc.PosY})
		}

	}
	e.livingCells = livingCellsNextGen
	e.deadCells = deadCellsNextGen
	e.generation++
}
