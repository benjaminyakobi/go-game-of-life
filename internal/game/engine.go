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
	livingCellsHistory cellsHistory
	patterns           *indexedPatterns
	patternName        string
}

type Engine interface {
	GetGeneration() int
	ResetGeneration()
	GetLivingCells() cellsSet
	SetLivingCells(cs cellsSet)
	PatternName() string
	SetPatternName(n string)
	Patterns() *indexedPatterns
	LivingCellsHistory() cellsHistory
	PreviousGeneration() cellsSet
	DeadCells() cellsSet
	DecrementGeneration()
	NextGeneration()
}

type cellsHistory struct {
	cells []cellsSet
	head  int
	count int
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

func initCellsHistory(size int) cellsHistory {
	return cellsHistory{
		cells: make([]cellsSet, size),
		head:  0,
		count: 0,
	}
}

func (h *cellsHistory) Len() int {
	return h.count
}

func (h *cellsHistory) Push(cs cellsSet) {
	h.cells[h.head] = cs
	h.head = (h.head + 1) % len(h.cells) // updating head position
	if h.count <= len(h.cells) {
		h.count++
	}
}

func (h *cellsHistory) Pop() cellsSet {
	if h.count == 0 {
		return nil
	}

	h.head = (h.head - 1 + len(h.cells)) % len(h.cells)

	cs := h.cells[h.head]
	h.cells[h.head] = nil
	h.count--

	return cs
}

// NOTE: wipe out the history
func (h *cellsHistory) Clear() {
	clear(h.cells)
	h.head = 0
	h.count = 0
}

// NOTE: updating for resize events that requires history updates also
func (h *cellsHistory) Update(update func(cellsSet) cellsSet) {
	for idx := range len(h.cells) {
		el := h.cells[idx]
		if el != nil {
			cs := update(el)
			h.cells[idx] = cs
		}
	}
}

func newEngine(cfg config) Engine {
	return &engine{
		generation:         0,
		deadCells:          make(cellsSet, 0),
		livingCells:        make(cellsSet, 0),
		livingCellsHistory: initCellsHistory(50),
		patterns:           cfg.loadPatterns(),
	}
}

func (e *engine) GetGeneration() int {
	return e.generation
}

func (e *engine) ResetGeneration() {
	e.generation = 0
}

func (e *engine) GetLivingCells() cellsSet {
	return e.livingCells
}

func (e *engine) SetLivingCells(cs cellsSet) {
	e.livingCells = cs
}

func (e *engine) PatternName() string {
	return e.patternName
}

func (e *engine) SetPatternName(n string) {
	e.patternName = n
}

func (e *engine) Patterns() *indexedPatterns {
	return e.patterns
}

func (e *engine) LivingCellsHistory() cellsHistory {
	return e.livingCellsHistory
}

func (e *engine) DeadCells() cellsSet {
	return e.deadCells
}

func (e *engine) DecrementGeneration() {
	e.generation--
}

func (e *engine) PreviousGeneration() cellsSet {
	lch := e.livingCellsHistory.Pop()
	return lch
}

func (e *engine) NextGeneration() {
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

	e.livingCellsHistory.Push(e.livingCells)

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
		// 0 living neighbors
		if _, exist := neighborCounts[lc]; !exist {
			deadCellsNextGen.Add(lc)
		}
	}

	for c, count := range neighborCounts {
		if count == 3 || e.livingCells.Contains(c) && count == 2 {
			livingCellsNextGen.Add(c)
		} else {
			deadCellsNextGen.Add(c)
		}
	}
	e.livingCells = livingCellsNextGen
	e.deadCells = deadCellsNextGen
	e.generation++
}
