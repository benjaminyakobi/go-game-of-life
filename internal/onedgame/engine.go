package onedgame

import (
	"slices"
)

type bitGrid struct {
	width    int
	height   int      // currently not in use
	grid     []uint64 // currently not in use
	living   []int
	gen      int
	patterns *flatIndexedPatterns
}

type BitEngine interface {
	nextGen() (flatCells, flatCells)
	getLiving() []int
	setLiving(i int)
	removeLiving(i int)
	resetGen()
	createIndex(x, y int) int
	revertIndex(i int) (int, int)
	getPattern(idx int) (string, flatCells)
}

func newBitEngine(cfg config, width, height int) (BitEngine, error) {
	size := width * height
	return &bitGrid{
		width:    width,
		height:   height,
		grid:     make([]uint64, (size+63)/64),
		living:   nil,
		gen:      0,
		patterns: cfg.loadPatterns(width),
	}, nil
}

func (bg *bitGrid) resetGen() {
	bg.gen = 0
}

func (bg *bitGrid) getLiving() []int {
	return bg.living
}

func (bg *bitGrid) setLiving(i int) {
	bg.living = append(bg.living, i)
}

func (bg *bitGrid) removeLiving(i int) {
	for j, v := range bg.living {
		if v == i {
			bg.living[j] = bg.living[len(bg.living)-1]
			bg.living = bg.living[:len(bg.living)-1]
			return
		}
	}
}

func (bg *bitGrid) nextGen() (flatCells, flatCells) {
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

	var livingNextGen flatCells
	var deadNextGen flatCells
	neighborCounter := make(map[int]int, len(bg.living)*8)

	for _, v := range bg.living {
		x, y := bg.revertIndex(v)
		for _, d := range directions {
			dx, dy := x+d[0], y+d[1]
			neighbor := bg.createIndex(dx, dy)
			neighborCounter[neighbor]++
		}
		// 0 living neighbors
		if _, exist := neighborCounter[v]; !exist {
			deadNextGen = append(deadNextGen, v)
		}
	}

	for c, count := range neighborCounter {
		if count == 3 || slices.Contains(bg.living, c) && count == 2 {
			livingNextGen = append(livingNextGen, c)
		} else {
			deadNextGen = append(deadNextGen, c)
		}
	}

	bg.living = livingNextGen
	return deadNextGen, livingNextGen
}

func (bg *bitGrid) createIndex(x, y int) int {
	return y*bg.width + x
}

func (bg *bitGrid) revertIndex(idx int) (int, int) {
	return idx % bg.width, idx / bg.width
}

func (bg *bitGrid) getPattern(idx int) (string, flatCells) {
	ip := bg.patterns
	cyclicIndex := idx % len(ip.keys)
	patternName := ip.keys[cyclicIndex]
	return patternName, ip.sets[patternName]
}
