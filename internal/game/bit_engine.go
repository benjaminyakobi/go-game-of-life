package game

import (
	"slices"
)

type bitGrid struct {
	width  int
	height int // currently not in use
	grid   []uint64
	living []int
	gen    int
}

type BitEngine interface {
	nextGen()
	getLiving() []int
	setLiving(i int)
	removeLiving(i int)
	resetGen()
	createIndex(x, y int) int
	revertIndex(i int) (int, int)
}

func newBitEngine(width, height int) (BitEngine, error) {
	size := width * height
	return &bitGrid{
		width:  width,
		height: height,
		grid:   make([]uint64, (size+63)/64),
		living: nil,
		gen:    0,
	}, nil
}

func (bg *bitGrid) resetGen() {
	bg.gen = 0
}

func (bg *bitGrid) getLiving() []int {
	return bg.living
}

func (bg *bitGrid) setLiving(i int) {
	// TODO: should be optimized with bg.grid setting to 1
	bg.living = append(bg.living, i)
}

func (bg *bitGrid) removeLiving(i int) {
	// TODO: should be optimized with bg.grid setting to 0
	for j, v := range bg.living {
		if v == i {
			copy(bg.living[j:], bg.living[j+1:])
			bg.living = bg.living[:len(bg.living)-1]
			return
		}
	}
}

func (bg *bitGrid) nextGen() {
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
	livingNextGen := []int{}
	neighborCounter := make(map[int]int, len(bg.living))

	for _, v := range bg.living {
		x, y := bg.revertIndex(v)

		for _, d := range directions {
			dx, dy := x+d[0], y+d[1]
			neighbor := bg.createIndex(dx, dy)
			neighborCounter[neighbor]++
		}
		// 0 living neighbors
		if _, exist := neighborCounter[v]; !exist {

		}

	}
	for c, count := range neighborCounter {
		if count == 3 || slices.Contains(bg.living, c) && count == 2 {
			livingNextGen = append(livingNextGen, c)
		} else {
		}
	}
	bg.living = livingNextGen
}

func (bg *bitGrid) createIndex(x, y int) int {
	return y*bg.width + x
}

func (bg *bitGrid) revertIndex(idx int) (int, int) {
	return idx % bg.width, idx / bg.width
}
