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
	gridWidth  int
	gridHeight int
	screen     tcell.Screen
	engine     *engine
}

var css = cellStyles{
	def:            tcell.StyleDefault.Background(color.Reset).Foreground(color.Default),
	lightSlateGrey: tcell.StyleDefault.Background(color.Reset).Foreground(color.LightSlateGrey),
	greenYellow:    tcell.StyleDefault.Background(color.Reset).Foreground(color.GreenYellow),
}

func initRenderer(e *engine) (*renderer, error) {
	// Screen must be initialized before we read its size.
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

	w, h := screen.Size() // usually int

	// tcell uses terminal cells.
	return &renderer{
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
	r.clearLine(y)
	textWidth := runewidth.StringWidth(text)

	calcX := (r.gridWidth - textWidth) / 2

	for row := range calcX {
		if row == 0 && y == screenOffset {
			r.screen.Put(row, y, string(tcell.RuneULCorner), css.def)
		} else if row > 0 && y == screenOffset {
			r.screen.Put(row, y, string(tcell.RuneHLine), css.def)
		} else if row == 0 && y == r.gridHeight-1 {
			r.screen.Put(row, y, string(tcell.RuneLLCorner), css.def)
		} else if row > 0 && y == r.gridHeight-1 {
			r.screen.Put(row, y, string(tcell.RuneHLine), css.def)
		}
	}

	col := calcX
	for _, row := range text {
		rw := runewidth.RuneWidth(row)
		r.screen.SetContent(col, y, row, nil, css.def)
		col += rw
	}

	for row := col; row < r.gridWidth; row++ {
		if row == r.gridWidth-1 && y == screenOffset {
			r.screen.Put(row, y, string(tcell.RuneURCorner), css.def)
		} else if row < r.gridWidth-1 && y == screenOffset {
			r.screen.Put(row, y, string(tcell.RuneHLine), css.def)
		} else if row == r.gridWidth-1 && y == r.gridHeight-1 {
			r.screen.Put(row, y, string(tcell.RuneLRCorner), css.def)
		} else if row < r.gridWidth-1 && y == r.gridHeight-1 {
			r.screen.Put(row, y, string(tcell.RuneHLine), css.def)
		}

	}
}

// TODO: should be reviewed, currently it's working fine
func (r *renderer) updateCellStyle(x, y int) {
	if y == screenOffset || y == r.gridHeight-1 {
		r.screen.Put(x, y, string(tcell.RuneHLine), css.def)
	} else if x == 0 || x == r.gridWidth-1 {
		r.screen.Put(x, y, string(tcell.RuneVLine), css.def)
	} else {
		r.screen.Put(x, y, ".", css.lightSlateGrey)
	}

	if x == 0 && y == screenOffset {
		r.screen.Put(x, y, string(tcell.RuneULCorner), css.def)
	} else if x == r.gridWidth-1 && y == screenOffset {
		r.screen.Put(x, y, string(tcell.RuneURCorner), css.def)
	} else if x == 0 && y == r.gridHeight-1 {
		r.screen.Put(x, y, string(tcell.RuneLLCorner), css.def)
	} else if x == r.gridWidth-1 && y == r.gridHeight-1 {
		r.screen.Put(x, y, string(tcell.RuneLRCorner), css.def)
	}
}

func (r *renderer) drawNewGrid() {
	r.drawText(0, "Click: Select | Double Click: Unselect | r: Run | p: Pause | s: Stop & Reset Generations | b: Clear & Choose Pattern | Left Arrow: Previous Generation | Right Arrow: Next Generation | =/-: Increase/Decrease Speed | Escapse: Exit")
	r.gridWidth, r.gridHeight = r.screen.Size()
	for w := range r.gridWidth {
		for h := screenOffset; h < r.gridHeight; h++ {
			r.updateCellStyle(w, h)
		}
	}
	r.drawText(1, gameText)
	r.drawText(r.gridHeight-1, "Conway's Game Of Life")
}

// TODO: should be reviewed, currently it's working fine
func (r *renderer) killLivingCellsOnGrid(cs cellsSet) {
	for c := range cs {
		if c.PosY > screenOffset && c.PosY < r.gridHeight-1 && c.PosX > 0 && c.PosX < r.gridWidth-1 {
			r.screen.Put(c.PosX, c.PosY, ".", css.lightSlateGrey)
		}
	}
}

func (r *renderer) drawLivingCellsOnGrid(cs cellsSet) {
	for c := range cs {
		if c.PosY > screenOffset && c.PosY < r.gridHeight-1 && c.PosX > 0 && c.PosX < r.gridWidth-1 {
			r.screen.Put(c.PosX, c.PosY, "@", css.greenYellow)
		}
	}
}

func (r *renderer) drawDeadCellsOnGrid(cs cellsSet) {
	for c := range cs {
		r.updateCellStyle(c.PosX, c.PosY)
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
	boxWidth, boxHeight, minWidth, minHeight = r.calcBoxDimesions()
	x := (r.gridWidth - boxWidth) / 2
	y := (r.gridHeight - boxHeight) / 2

	for col := x; col < x+boxWidth; col++ {
		r.screen.Put(col, y, ".", css.lightSlateGrey)
		r.screen.Put(col, y+boxHeight-1, ".", css.lightSlateGrey)
		// r.screen.SetContent(col, y+boxHeight-1, tcell.RuneHLine, nil, cs.def)
	}

	for row := y; row < y+boxHeight; row++ {
		r.screen.Put(x, row, ".", css.lightSlateGrey)
		r.screen.Put(x+boxWidth-1, row, ".", css.lightSlateGrey)
		// r.screen.SetContent(x+boxWidth-1, row, tcell.RuneVLine, nil, cs.def)
	}

	// r.screen.SetContent(x, y, tcell.RuneULCorner, nil, cs.def)
	// r.screen.SetContent(x+boxWidth-1, y, tcell.RuneURCorner, nil, cs.def)
	// r.screen.SetContent(x, y+boxHeight-1, tcell.RuneLLCorner, nil, cs.def)
	// r.screen.SetContent(x+boxWidth-1, y+boxHeight-1, tcell.RuneLRCorner, nil, cs.def)

	// centerLivingCells := func(lcs livingCellsSet) livingCellsSet { centeredLCS := make(livingCellsSet)
	// 	for cell := range lcs {
	// 		PosX := x + cell.PosX - minWidth + 2
	// 		PosY := y + cell.PosY - minHeight + 2
	// 		centeredLCS.Add(livingCell{PosX: PosX, PosY: PosY})
	// 	}
	// 	return centeredLCS
	// }
	//
	// r.engine.livingCells = centerLivingCells(r.engine.livingCells)
	// r.drawLivingCellsOnGrid()
	// r.drawText(1, title)
}

func (r *renderer) drawBox(title string) {
	boxWidth, boxHeight, minWidth, minHeight = r.calcBoxDimesions()
	x := (r.gridWidth - boxWidth) / 2
	y := (r.gridHeight - boxHeight) / 2

	for col := x; col < x+boxWidth; col++ {
		r.screen.SetContent(col, y, tcell.RuneHLine, nil, css.def)
		r.screen.SetContent(col, y+boxHeight-1, tcell.RuneHLine, nil, css.def)
	}

	for row := y; row < y+boxHeight; row++ {
		r.screen.SetContent(x, row, tcell.RuneVLine, nil, css.def)
		r.screen.SetContent(x+boxWidth-1, row, tcell.RuneVLine, nil, css.def)
	}

	r.screen.SetContent(x, y, tcell.RuneULCorner, nil, css.def)
	r.screen.SetContent(x+boxWidth-1, y, tcell.RuneURCorner, nil, css.def)
	r.screen.SetContent(x, y+boxHeight-1, tcell.RuneLLCorner, nil, css.def)
	r.screen.SetContent(x+boxWidth-1, y+boxHeight-1, tcell.RuneLRCorner, nil, css.def)

	// TODO: make it separate func to be reused also when resizing the grid!
	centerLivingCells := func(cs cellsSet) cellsSet {
		centeredCS := make(cellsSet)
		for c := range cs {
			PosX := x + c.PosX - minWidth + 2
			PosY := y + c.PosY - minHeight + 2
			centeredCS.Add(cell{PosX: PosX, PosY: PosY})
		}
		return centeredCS
	}

	r.engine.livingCells = centerLivingCells(r.engine.livingCells)
	r.drawLivingCellsOnGrid(r.engine.livingCells)
	r.drawText(1, title)
}
