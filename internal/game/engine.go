package game

// Contains game engine code

type cell struct {
	PosX int `json:"x"`
	PosY int `json:"y"`
}

type cellsSet map[cell]struct{}

type engine struct {
	generation         int
	deadCells          cellsSet
	livingCells        cellsSet
	livingCellsHistory []cellsSet
	historySize        int
	patterns           []cellsSet
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

func initEngine(cfg config) *engine {
	return &engine{
		generation:         0,
		deadCells:          make(cellsSet, 0),
		livingCells:        make(cellsSet, 0),
		livingCellsHistory: make([]cellsSet, 0),
		historySize:        50,
		patterns:           cfg.loadPatterns(),
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

	if len(e.livingCellsHistory) >= e.historySize {
		e.livingCellsHistory = e.livingCellsHistory[1:]
	}
	e.livingCellsHistory = append(e.livingCellsHistory, e.livingCells)

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
