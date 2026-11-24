package bitwise

type grid struct {
	width int
	grid  []uint64
}

type Engine interface{}

func newEngine(w, h int) (Engine, error) {
	size := w * h
	return &grid{
		width: w,
		grid:  make([]uint64, (size+63)/64),
	}, nil
}
