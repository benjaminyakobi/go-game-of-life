package game

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

func (bg *bitGrid) resetGen() {
	bg.gen = 0
}

func (bg *bitGrid) createIndex(x, y int) int {
	return y*bg.width + x
}

func (bg *bitGrid) revertIndex(idx int) (int, int) {
	return idx % bg.width, idx / bg.width
}
