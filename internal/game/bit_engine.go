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
	// TODO: implement
}

func (bg *bitGrid) createIndex(x, y int) int {
	return y*bg.width + x
}

func (bg *bitGrid) revertIndex(idx int) (int, int) {
	return idx % bg.width, idx / bg.width
}
