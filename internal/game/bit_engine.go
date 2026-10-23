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
}

func (bg *bitGrid) createIndex(x, y int) int {
	return y*bg.width + x
}

func (bg *bitGrid) revertIndex(idx int) (int, int) {
	return idx % bg.width, idx / bg.width
}
