package game

// Contains game grid code

import (
	"math"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/mattn/go-runewidth"
)

type CellStyles struct {
	def            tcell.Style
	grey           tcell.Style
	lightSlateGrey tcell.Style
	greenYellow    tcell.Style
}

var cellStyles = CellStyles{
	def:            tcell.StyleDefault.Background(color.Reset).Foreground(color.Default),
	lightSlateGrey: tcell.StyleDefault.Background(color.Reset).Foreground(color.LightSlateGrey),
	greenYellow:    tcell.StyleDefault.Background(color.Reset).Foreground(color.GreenYellow),
}

func clearLine(s tcell.Screen, y int) {
	w, _ := s.Size()
	for x := range w {
		s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
	}
}

func drawText(s tcell.Screen, y int, text string) {
	clearLine(s, y)
	w, h := s.Size()
	textWidth := runewidth.StringWidth(text)

	calcX := (w - textWidth) / 2

	for r := range calcX {
		if r == 0 && y == screenOffset {
			s.Put(r, y, string(tcell.RuneULCorner), cellStyles.def)
		} else if r > 0 && y == screenOffset {
			s.Put(r, y, string(tcell.RuneHLine), cellStyles.def)
		} else if r == 0 && y == h-1 {
			s.Put(r, y, string(tcell.RuneLLCorner), cellStyles.def)
		} else if r > 0 && y == h-1 {
			s.Put(r, y, string(tcell.RuneHLine), cellStyles.def)
		}
	}

	col := calcX
	for _, r := range text {
		rw := runewidth.RuneWidth(r)
		s.SetContent(col, y, r, nil, cellStyles.def)
		col += rw
	}

	for r := col; r < w; r++ {
		if r == w-1 && y == screenOffset {
			s.Put(r, y, string(tcell.RuneURCorner), cellStyles.def)
		} else if r < w-1 && y == screenOffset {
			s.Put(r, y, string(tcell.RuneHLine), cellStyles.def)
		} else if r == w-1 && y == h-1 {
			s.Put(r, y, string(tcell.RuneLRCorner), cellStyles.def)
		} else if r < w-1 && y == h-1 {
			s.Put(r, y, string(tcell.RuneHLine), cellStyles.def)
		}

	}
}

func updateCellStyle(s tcell.Screen, x, y int) {
	w, h := s.Size()
	if y == screenOffset || y == h-1 {
		s.Put(x, y, string(tcell.RuneHLine), cellStyles.def)
	} else if x == 0 || x == w-1 {
		s.Put(x, y, string(tcell.RuneVLine), cellStyles.def)
	} else {
		s.Put(x, y, ".", cellStyles.lightSlateGrey)
	}

	if x == 0 && y == screenOffset {
		s.Put(x, y, string(tcell.RuneULCorner), cellStyles.def)
	} else if x == w-1 && y == screenOffset {
		s.Put(x, y, string(tcell.RuneURCorner), cellStyles.def)
	} else if x == 0 && y == h-1 {
		s.Put(x, y, string(tcell.RuneLLCorner), cellStyles.def)
	} else if x == w-1 && y == h-1 {
		s.Put(x, y, string(tcell.RuneLRCorner), cellStyles.def)
	}
}

func drawNewGrid(s tcell.Screen) {
	_, h := s.Size()
	drawText(s, 0, "Click: Select | Double Click: Unselect | r: Run | p: Pause | s: Stop & Reset Generations | b: Clear & Choose Pattern | Left Arrow: Previous Generation | Right Arrow: Next Generation | =/-: Increase/Decrease Speed | Escapse: Exit")
	width, height := s.Size()
	for w := range width {
		for h := screenOffset; h < height; h++ {
			updateCellStyle(s, w, h)
		}
	}
	drawText(s, 1, gameText)
	drawText(s, h-1, "Conway's Game Of Life")
}

func drawLivingCellsOnGrid(s tcell.Screen) {
	w, h := s.Size()
	for lc := range livingCells {
		if lc.PosY > screenOffset && lc.PosY < h-1 && lc.PosX > 0 && lc.PosX < w-1 {
			s.Put(lc.PosX, lc.PosY, "@", cellStyles.greenYellow)
		}
	}
}

func calcBoxDimesions(s tcell.Screen) (int, int, int, int) {
	sw, sh := s.Size()
	minW, maxW := sw, math.MinInt32
	minH, maxH := sh, math.MinInt32
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

func drawBox(s tcell.Screen, title string) {
	boxWidth, boxHeight, minWidth, minHeight = calcBoxDimesions(s)
	sw, sh := s.Size()
	x := (sw - boxWidth) / 2
	y := (sh - boxHeight) / 2

	for col := x; col < x+boxWidth; col++ {
		s.SetContent(col, y, tcell.RuneHLine, nil, cellStyles.def)
		s.SetContent(col, y+boxHeight-1, tcell.RuneHLine, nil, cellStyles.def)
	}

	for row := y; row < y+boxHeight; row++ {
		s.SetContent(x, row, tcell.RuneVLine, nil, cellStyles.def)
		s.SetContent(x+boxWidth-1, row, tcell.RuneVLine, nil, cellStyles.def)
	}

	s.SetContent(x, y, tcell.RuneULCorner, nil, cellStyles.def)
	s.SetContent(x+boxWidth-1, y, tcell.RuneURCorner, nil, cellStyles.def)
	s.SetContent(x, y+boxHeight-1, tcell.RuneLLCorner, nil, cellStyles.def)
	s.SetContent(x+boxWidth-1, y+boxHeight-1, tcell.RuneLRCorner, nil, cellStyles.def)

	centerLivingCells := func(lcs LivingCellsSet) LivingCellsSet {
		centeredLCS := make(LivingCellsSet)
		for cell := range lcs {
			PosX := x + cell.PosX - minWidth + 2
			PosY := y + cell.PosY - minHeight + 2
			centeredLCS.Add(LivingCell{PosX: PosX, PosY: PosY})
		}
		return centeredLCS
	}

	livingCells = centerLivingCells(livingCells)
	drawLivingCellsOnGrid(s)
	drawText(s, 1, title)
}
