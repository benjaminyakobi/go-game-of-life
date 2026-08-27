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
}

var cs = cellStyles{
	def:            tcell.StyleDefault.Background(color.Reset).Foreground(color.Default),
	lightSlateGrey: tcell.StyleDefault.Background(color.Reset).Foreground(color.LightSlateGrey),
	greenYellow:    tcell.StyleDefault.Background(color.Reset).Foreground(color.GreenYellow),
}

func initRenderer() (*renderer, error) {
	// Screen must be initialized before we read its size.
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.SetStyle(cs.def)
	screen.EnableMouse()
	screen.Clear()

	w, h := screen.Size() // usually int

	// tcell uses terminal cells.
	return &renderer{
		gridWidth:  w,
		gridHeight: h,
		screen:     screen,
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
			r.screen.Put(row, y, string(tcell.RuneULCorner), cs.def)
		} else if row > 0 && y == screenOffset {
			r.screen.Put(row, y, string(tcell.RuneHLine), cs.def)
		} else if row == 0 && y == r.gridHeight-1 {
			r.screen.Put(row, y, string(tcell.RuneLLCorner), cs.def)
		} else if row > 0 && y == r.gridHeight-1 {
			r.screen.Put(row, y, string(tcell.RuneHLine), cs.def)
		}
	}

	col := calcX
	for _, row := range text {
		rw := runewidth.RuneWidth(row)
		r.screen.SetContent(col, y, row, nil, cs.def)
		col += rw
	}

	for row := col; row < r.gridWidth; row++ {
		if row == r.gridWidth-1 && y == screenOffset {
			r.screen.Put(row, y, string(tcell.RuneURCorner), cs.def)
		} else if row < r.gridWidth-1 && y == screenOffset {
			r.screen.Put(row, y, string(tcell.RuneHLine), cs.def)
		} else if row == r.gridWidth-1 && y == r.gridHeight-1 {
			r.screen.Put(row, y, string(tcell.RuneLRCorner), cs.def)
		} else if row < r.gridWidth-1 && y == r.gridHeight-1 {
			r.screen.Put(row, y, string(tcell.RuneHLine), cs.def)
		}

	}
}

func (r *renderer) updateCellStyle(x, y int) {
	if y == screenOffset || y == r.gridHeight-1 {
		r.screen.Put(x, y, string(tcell.RuneHLine), cs.def)
	} else if x == 0 || x == r.gridWidth-1 {
		r.screen.Put(x, y, string(tcell.RuneVLine), cs.def)
	} else {
		r.screen.Put(x, y, ".", cs.lightSlateGrey)
	}

	if x == 0 && y == screenOffset {
		r.screen.Put(x, y, string(tcell.RuneULCorner), cs.def)
	} else if x == r.gridWidth-1 && y == screenOffset {
		r.screen.Put(x, y, string(tcell.RuneURCorner), cs.def)
	} else if x == 0 && y == r.gridHeight-1 {
		r.screen.Put(x, y, string(tcell.RuneLLCorner), cs.def)
	} else if x == r.gridWidth-1 && y == r.gridHeight-1 {
		r.screen.Put(x, y, string(tcell.RuneLRCorner), cs.def)
	}
}

func (r *renderer) drawNewGrid() {
	r.drawText(0, "Click: Select | Double Click: Unselect | r: Run | p: Pause | s: Stop & Reset Generations | b: Clear & Choose Pattern | Left Arrow: Previous Generation | Right Arrow: Next Generation | =/-: Increase/Decrease Speed | Escapse: Exit")
	for w := range r.gridWidth {
		for h := screenOffset; h < r.gridHeight; h++ {
			r.updateCellStyle(w, h)
		}
	}
	r.drawText(1, gameText)
	r.drawText(r.gridHeight-1, "Conway's Game Of Life")
}

func (r *renderer) drawLivingCellsOnGrid() {
	for lc := range livingCells {
		if lc.PosY > screenOffset && lc.PosY < r.gridHeight-1 && lc.PosX > 0 && lc.PosX < r.gridWidth-1 {
			r.screen.Put(lc.PosX, lc.PosY, "@", cs.greenYellow)
		}
	}
}

func (r *renderer) calcBoxDimesions() (int, int, int, int) {
	minW, maxW := r.gridWidth, math.MinInt32
	minH, maxH := r.gridHeight, math.MinInt32
	for cell := range livingCells {
		minW = min(minW, cell.PosX)
		maxW = max(maxW, cell.PosX)
		minH = min(minH, cell.PosY)
		maxH = max(maxH, cell.PosY)
	}
	if livingCells.Len() == 1 {
		return 5, 5, minW, minH
	}
	return maxW - minW + 5, maxH - minH + 5, minW, minH
}

func (r *renderer) drawBox(title string) {
	boxWidth, boxHeight, minWidth, minHeight = r.calcBoxDimesions()
	x := (r.gridWidth - boxWidth) / 2
	y := (r.gridHeight - boxHeight) / 2

	for col := x; col < x+boxWidth; col++ {
		r.screen.SetContent(col, y, tcell.RuneHLine, nil, cs.def)
		r.screen.SetContent(col, y+boxHeight-1, tcell.RuneHLine, nil, cs.def)
	}

	for row := y; row < y+boxHeight; row++ {
		r.screen.SetContent(x, row, tcell.RuneVLine, nil, cs.def)
		r.screen.SetContent(x+boxWidth-1, row, tcell.RuneVLine, nil, cs.def)
	}

	r.screen.SetContent(x, y, tcell.RuneULCorner, nil, cs.def)
	r.screen.SetContent(x+boxWidth-1, y, tcell.RuneURCorner, nil, cs.def)
	r.screen.SetContent(x, y+boxHeight-1, tcell.RuneLLCorner, nil, cs.def)
	r.screen.SetContent(x+boxWidth-1, y+boxHeight-1, tcell.RuneLRCorner, nil, cs.def)

	centerLivingCells := func(lcs livingCellsSet) livingCellsSet {
		centeredLCS := make(livingCellsSet)
		for cell := range lcs {
			PosX := x + cell.PosX - minWidth + 2
			PosY := y + cell.PosY - minHeight + 2
			centeredLCS.Add(livingCell{PosX: PosX, PosY: PosY})
		}
		return centeredLCS
	}

	livingCells = centerLivingCells(livingCells)
	r.drawLivingCellsOnGrid()
	r.drawText(1, title)
}
