package game

// Contains game grid code

import (
	"math"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/mattn/go-runewidth"
)

type cellStyles struct {
	def            tcell.Style
	grey           tcell.Style
	lightSlateGrey tcell.Style
	greenYellow    tcell.Style
}

type renderer struct {
	gridOffset int
	gridWidth  int
	gridHeight int
	boxWidth   int
	boxHeight  int
	screen     tcell.Screen
	engine     *engine
}

var css = cellStyles{
	def:            tcell.StyleDefault.Background(color.Reset).Foreground(color.Default),
	lightSlateGrey: tcell.StyleDefault.Background(color.Reset).Foreground(color.LightSlateGrey),
	greenYellow:    tcell.StyleDefault.Background(color.Reset).Foreground(color.GreenYellow),
}

func initRenderer(e *engine) (*renderer, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.SetStyle(css.def)
	screen.EnableMouse()
	screen.Clear()

	w, h := screen.Size()

	return &renderer{
		gridOffset: 1,
		gridWidth:  w,
		gridHeight: h,
		screen:     screen,
		engine:     e,
	}, nil
}

func (r *renderer) clearLine(y int) {
	for x := range r.gridWidth {
		r.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
	}
}

func (r *renderer) drawText(y int, text string) {
	r.clearLine(y) // 1. clear line

	// 2. calculate test position
	textWidth := runewidth.StringWidth(text)
	startX := max(1, (r.gridWidth-textWidth)/2)
	endX := startX

	// 3. draw text
	for _, ch := range text {
		rw := runewidth.RuneWidth(ch)
		r.screen.SetContent(endX, y, ch, nil, css.def)
		endX += rw
	}

	// 4. is this a border row (top/bottom)? return if no
	if y != r.gridOffset && y != r.gridHeight-1 {
		return
	}

	// 5. draw left/right corners
	var leftCorner, rightCorner rune

	if y == r.gridOffset { // top screen
		leftCorner = tcell.RuneULCorner
		rightCorner = tcell.RuneURCorner
	} else { // bottom screen
		leftCorner = tcell.RuneLLCorner
		rightCorner = tcell.RuneLRCorner
	}

	r.screen.Put(0, y, string(leftCorner), css.def)
	r.screen.Put(r.gridWidth-1, y, string(rightCorner), css.def)

	// 6. fill horizontal lines
	for x := 1; x < startX; x++ {
		r.screen.Put(x, y, string(tcell.RuneHLine), css.def)
	}

	for x := endX; x < r.gridWidth-1; x++ {
		r.screen.Put(x, y, string(tcell.RuneHLine), css.def)
	}
}

func (r *renderer) updateCellStyle(x, y int) {
	switch {
	case x == 0 && y == r.gridOffset:
		r.screen.Put(x, y, string(tcell.RuneULCorner), css.def)

	case x == r.gridWidth-1 && y == r.gridOffset:
		r.screen.Put(x, y, string(tcell.RuneURCorner), css.def)

	case x == 0 && y == r.gridHeight-1:
		r.screen.Put(x, y, string(tcell.RuneLLCorner), css.def)

	case x == r.gridWidth-1 && y == r.gridHeight-1:
		r.screen.Put(x, y, string(tcell.RuneLRCorner), css.def)

	case y == r.gridOffset || y == r.gridHeight-1:
		r.screen.Put(x, y, string(tcell.RuneHLine), css.def)

	case x == 0 || x == r.gridWidth-1:
		r.screen.Put(x, y, string(tcell.RuneVLine), css.def)

	default:
		r.drawSingleDeadCellOnGrid(cell{PosX: x, PosY: y})
	}
}

func (r *renderer) drawNewGrid() {
	r.drawText(0, "Click: Select | Double Click: Unselect | r: Run | p: Pause | s: Stop & Reset Generations | b: Clear & Choose Pattern | Left Arrow: Previous Generation | Right Arrow: Next Generation | =/-: Increase/Decrease Speed | Escapse: Exit")
	r.gridWidth, r.gridHeight = r.screen.Size()
	for w := range r.gridWidth {
		for h := r.gridOffset; h < r.gridHeight; h++ {
			r.updateCellStyle(w, h)
		}
	}
	r.drawText(1, "")
	r.drawText(r.gridHeight-1, "Conway's Game Of Life")
}

func (r *renderer) killLivingCellsOnGrid(cs cellsSet) {
	for c := range cs {
		r.drawSingleDeadCellOnGrid(c)
	}
}

func (r *renderer) drawSingleDeadCellOnGrid(c cell) {
	if c.PosY > r.gridOffset &&
		c.PosY < r.gridHeight-1 &&
		c.PosX > 0 &&
		c.PosX < r.gridWidth-1 {
		r.screen.Put(c.PosX, c.PosY, ".", css.lightSlateGrey)
	}
}

func (r *renderer) drawSingleLivingCellOnGrid(c cell) {
	if c.PosY > r.gridOffset &&
		c.PosY < r.gridHeight-1 &&
		c.PosX > 0 &&
		c.PosX < r.gridWidth-1 {
		r.screen.Put(c.PosX, c.PosY, "@", css.greenYellow)
	}
}

func (r *renderer) drawLivingCellsOnGrid(cs cellsSet) {
	for c := range cs {
		r.drawSingleLivingCellOnGrid(c)
	}
}

func (r *renderer) drawDeadCellsOnGrid(cs cellsSet) {
	for c := range cs {
		r.drawSingleDeadCellOnGrid(c)
	}
}

func (r *renderer) calcBoxDimesions() (int, int, int, int) {
	minW, maxW := r.gridWidth, math.MinInt32
	minH, maxH := r.gridHeight, math.MinInt32
	for cell := range r.engine.livingCells {
		minW = min(minW, cell.PosX)
		maxW = max(maxW, cell.PosX)
		minH = min(minH, cell.PosY)
		maxH = max(maxH, cell.PosY)
	}
	if r.engine.livingCells.Len() == 1 {
		return 5, 5, minW, minH
	}
	return maxW - minW + 5, maxH - minH + 5, minW, minH
}

func (r *renderer) removeBox() {
	if r.boxWidth == -1 && r.boxHeight == -1 {
		return
	}

	x := (r.gridWidth - r.boxWidth) / 2
	y := (r.gridHeight - r.boxHeight) / 2

	for col := x; col < x+r.boxWidth; col++ {
		r.screen.Put(col, y, ".", css.lightSlateGrey)
		r.screen.Put(col, y+r.boxHeight-1, ".", css.lightSlateGrey)
	}

	for row := y; row < y+r.boxHeight; row++ {
		r.screen.Put(x, row, ".", css.lightSlateGrey)
		r.screen.Put(x+r.boxWidth-1, row, ".", css.lightSlateGrey)
	}

	r.boxWidth, r.boxHeight = -1, -1
}

func (r *renderer) centerSliceOfCells(css []cellsSet) []cellsSet {
	centered := make([]cellsSet, 0)
	for _, cs := range css {
		cc := r.centerCells(cs,
			0, 0, r.gridWidth, r.gridHeight)
		centered = append(centered, cc)
	}
	return centered
}

func (r *renderer) centerCells(
	cs cellsSet,
	x, y, width, height int,
) cellsSet {
	if len(cs) == 0 {
		return make(cellsSet)
	}

	minX := math.MaxInt
	minY := math.MaxInt
	maxX := math.MinInt
	maxY := math.MinInt

	for c := range cs {
		minX = min(minX, c.PosX)
		minY = min(minY, c.PosY)
		maxX = max(maxX, c.PosX)
		maxY = max(maxY, c.PosY)
	}

	patternWidth := maxX - minX + 1
	patternHeight := maxY - minY + 1

	offsetX := x + (width-patternWidth)/2
	offsetY := y + (height-patternHeight)/2

	centeredCS := make(cellsSet)

	for c := range cs {
		centeredCS.Add(cell{
			PosX: offsetX + (c.PosX - minX),
			PosY: offsetY + (c.PosY - minY),
		})
	}

	return centeredCS
}

func (r *renderer) drawBox(title string) {
	r.boxWidth, r.boxHeight, _, _ = r.calcBoxDimesions()

	x := (r.gridWidth - r.boxWidth) / 2
	y := (r.gridHeight - r.boxHeight) / 2

	// Top and bottom borders.
	for col := x; col < x+r.boxWidth; col++ {
		r.screen.SetContent(
			col,
			y,
			tcell.RuneHLine,
			nil,
			css.def,
		)

		r.screen.SetContent(
			col,
			y+r.boxHeight-1,
			tcell.RuneHLine,
			nil,
			css.def,
		)
	}

	// Left and right borders.
	for row := y; row < y+r.boxHeight; row++ {
		r.screen.SetContent(
			x,
			row,
			tcell.RuneVLine,
			nil,
			css.def,
		)

		r.screen.SetContent(
			x+r.boxWidth-1,
			row,
			tcell.RuneVLine,
			nil,
			css.def,
		)
	}

	// Corners.
	r.screen.SetContent(
		x,
		y,
		tcell.RuneULCorner,
		nil,
		css.def,
	)

	r.screen.SetContent(
		x+r.boxWidth-1,
		y,
		tcell.RuneURCorner,
		nil,
		css.def,
	)

	r.screen.SetContent(
		x,
		y+r.boxHeight-1,
		tcell.RuneLLCorner,
		nil,
		css.def,
	)

	r.screen.SetContent(
		x+r.boxWidth-1,
		y+r.boxHeight-1,
		tcell.RuneLRCorner,
		nil,
		css.def,
	)

	// Center the current pattern inside the box.
	r.engine.livingCells = r.centerCells(
		r.engine.livingCells,
		x+1,
		y+1,
		r.boxWidth-2,
		r.boxHeight-2,
	)

	r.drawLivingCellsOnGrid(r.engine.livingCells)
	r.drawText(1, title)
}
