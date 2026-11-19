package onedgame

import (
	"slices"
)

type grid struct {
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
	initLiving(living flatCells)
	removeLiving(i int)
	resetGen()
	createIndex(x, y int) int
	revertIndex(i int) (int, int)
	getPattern(idx int) (string, flatCells)
}

func newBitEngine(cfg config, width, height int) (BitEngine, error) {
	size := width * height
	return &grid{
		width:    width,
		height:   height,
		grid:     make([]uint64, (size+63)/64),
		living:   nil,
		gen:      0,
		patterns: cfg.loadPatterns(width),
	}, nil
}

func (g *grid) resetGen() {
	g.gen = 0
}

func (g *grid) getLiving() []int {
	return g.living
}

func (g *grid) setLiving(i int) {
	g.living = append(g.living, i)
}

func (g *grid) initLiving(living flatCells) {
	g.living = living
}

func (g *grid) removeLiving(i int) {
	for j, v := range g.living {
		if v == i {
			g.living[j] = g.living[len(g.living)-1]
			g.living = g.living[:len(g.living)-1]
			return
		}
	}
}

func (g *grid) nextGen() (flatCells, flatCells) {
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
	neighborCounter := make(map[int]int, len(g.living)*8)

	for _, v := range g.living {
		x, y := g.revertIndex(v)
		for _, d := range directions {
			dx, dy := x+d[0], y+d[1]
			neighbor := g.createIndex(dx, dy)
			neighborCounter[neighbor]++
		}
		// 0 living neighbors
		if _, exist := neighborCounter[v]; !exist {
			deadNextGen = append(deadNextGen, v)
		}
	}

	for c, count := range neighborCounter {
		if count == 3 || slices.Contains(g.living, c) && count == 2 {
			livingNextGen = append(livingNextGen, c)
		} else {
			deadNextGen = append(deadNextGen, c)
		}
	}

	g.living = livingNextGen
	return deadNextGen, livingNextGen
}

func (g *grid) createIndex(x, y int) int {
	return y*g.width + x
}

func (g *grid) revertIndex(idx int) (int, int) {
	return idx % g.width, idx / g.width
}

func (g *grid) getPattern(idx int) (string, flatCells) {
	ip := g.patterns
	cyclicIndex := idx % len(ip.keys)
	patternName := ip.keys[cyclicIndex]
	return patternName, ip.sets[patternName]
}
