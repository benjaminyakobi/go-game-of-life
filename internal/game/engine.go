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

func (e *engine) calcNextGeneration() {
	directions := [][]int{
		{-1, -1}, // top left
		{0, -1},  // top mid
		{1, -1},  // top right
		{-1, 0},  // left
		{1, 0},   // right
		{-1, 1},  // bottom left
		{0, 1},   // bottom mid
		{1, 1},   // bottom right
	}

	if e.livingCellsHistory.Len() > historySize {
		e.livingCellsHistory.Remove(e.livingCellsHistory.Front())
	}
	e.livingCellsHistory.PushBack(e.livingCells)

	livingCellsNextGen := make(cellsSet)
	deadCellsNextGen := make(cellsSet)
	neighborCounts := make(map[cell]int, len(e.livingCells)*8)

	for lc := range e.livingCells {
		for _, d := range directions {
			dx, dy := d[0], d[1]
			neighborCell := cell{
				PosX: lc.PosX + dx,
				PosY: lc.PosY + dy,
			}
			neighborCounts[neighborCell]++
		}
	}

	for c, count := range neighborCounts {
		if e.livingCells.Contains(c) {
			if count == 2 || count == 3 {
				livingCellsNextGen.Add(c)
			} else {
				deadCellsNextGen.Add(c)
			}
		} else if count == 3 {
			livingCellsNextGen.Add(c)
		}
	}

	e.livingCells = livingCellsNextGen
	e.deadCells = deadCellsNextGen
	e.generation++
}
