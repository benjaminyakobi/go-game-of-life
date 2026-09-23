package game

// Contains game grid code

import (
	"math"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
	"github.com/mattn/go-runewidth"
)

type cellStyles struct {
	def    tcell.Style
	border tcell.Style
	hud    tcell.Style
	dead   tcell.Style
	live   tcell.Style
	box    tcell.Style
}

const (
	deadCellGlyph   = "·"
	livingCellGlyph = "◆"
)

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
	def:    tcell.StyleDefault.Background(color.Black).Foreground(color.LightCyan),
	border: tcell.StyleDefault.Background(color.Black).Foreground(color.MediumOrchid).Bold(true),
	hud:    tcell.StyleDefault.Background(color.Black).Foreground(color.HotPink).Bold(true),
	dead:   tcell.StyleDefault.Background(color.Black).Foreground(color.DarkSlateBlue),
	live:   tcell.StyleDefault.Background(color.Black).Foreground(color.Aqua).Bold(true),
	box:    tcell.StyleDefault.Background(color.Black).Foreground(color.DeepSkyBlue).Bold(true),
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
		r.screen.SetContent(x, y, ' ', nil, css.def)
	}
}

func (r *renderer) drawText(y int, text string) {
	r.clearLine(y) // 1. clear line

	// 2. calculate test position
	startX := 1
	if y > 0 {
		textWidth := runewidth.StringWidth(text)
		startX = max(1, (r.gridWidth-textWidth)/2)
	}
	endX := startX

	// 3. draw text
	for _, ch := range text {
		rw := runewidth.RuneWidth(ch)
		r.screen.SetContent(endX, y, ch, nil, css.hud)
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

	r.screen.Put(0, y, string(leftCorner), css.border)
	r.screen.Put(r.gridWidth-1, y, string(rightCorner), css.border)

	// 6. fill horizontal lines
	for x := 1; x < startX; x++ {
		r.screen.Put(x, y, string(tcell.RuneHLine), css.border)
	}

	for x := endX; x < r.gridWidth-1; x++ {
		r.screen.Put(x, y, string(tcell.RuneHLine), css.border)
	}
}

func (r *renderer) updateCellStyle(x, y int) {
	switch {
	case x == 0 && y == r.gridOffset:
		r.screen.Put(x, y, string(tcell.RuneULCorner), css.border)

	case x == r.gridWidth-1 && y == r.gridOffset:
		r.screen.Put(x, y, string(tcell.RuneURCorner), css.border)

	case x == 0 && y == r.gridHeight-1:
		r.screen.Put(x, y, string(tcell.RuneLLCorner), css.border)

	case x == r.gridWidth-1 && y == r.gridHeight-1:
		r.screen.Put(x, y, string(tcell.RuneLRCorner), css.border)

	case y == r.gridOffset || y == r.gridHeight-1:
		r.screen.Put(x, y, string(tcell.RuneHLine), css.border)

	case x == 0 || x == r.gridWidth-1:
		r.screen.Put(x, y, string(tcell.RuneVLine), css.border)

	default:
		r.drawSingleDeadCellOnGrid(cell{PosX: x, PosY: y})
	}
}

func (r *renderer) drawNewGrid() {
	r.drawText(0, "click select | dbl-click clear | r run | p pause | s stop | b pattern | ←/→ history | +/- speed | esc exit")
	r.gridWidth, r.gridHeight = r.screen.Size()
	for w := range r.gridWidth {
		for h := r.gridOffset; h < r.gridHeight; h++ {
			r.updateCellStyle(w, h)
		}
	}
	r.drawText(1, "")
	r.drawText(r.gridHeight-1, "NEON LIFE")
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
		r.screen.Put(c.PosX, c.PosY, deadCellGlyph, css.dead)
	}
}

func (r *renderer) drawSingleLivingCellOnGrid(c cell) {
	if c.PosY > r.gridOffset &&
		c.PosY < r.gridHeight-1 &&
		c.PosX > 0 &&
		c.PosX < r.gridWidth-1 {
		r.screen.Put(c.PosX, c.PosY, livingCellGlyph, css.live)
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

func (r *renderer) calcPatternDimesions(cs cellsSet) (
	int, int, int, int, int, int) {
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

	return patternWidth, patternHeight, minX, minY, maxX, maxY
}

func (r *renderer) removeBox() {
	if r.boxWidth == -1 && r.boxHeight == -1 {
		return
	}

	x := (r.gridWidth - r.boxWidth) / 2
	y := (r.gridHeight - r.boxHeight) / 2

	for col := x; col < x+r.boxWidth; col++ {
		r.screen.Put(col, y, deadCellGlyph, css.dead)
		r.screen.Put(col, y+r.boxHeight-1, deadCellGlyph, css.dead)
	}

	for row := y; row < y+r.boxHeight; row++ {
		r.screen.Put(x, row, deadCellGlyph, css.dead)
		r.screen.Put(x+r.boxWidth-1, row, deadCellGlyph, css.dead)
	}

	r.boxWidth, r.boxHeight = -1, -1
}

func (r *renderer) centerCellsHistory(history *cellsHistory) {
	history.Update(func(cs cellsSet) cellsSet {
		return r.centerCells(cs, 0, 0, r.gridWidth, r.gridHeight)
	})
}

func (r *renderer) centerCells(
	cs cellsSet,
	x, y, width, height int,
) cellsSet {
	if len(cs) == 0 {
		return make(cellsSet)
	}

	patternWidth, patternHeight, minX, minY, _, _ := r.calcPatternDimesions(cs)

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
	r.boxWidth, r.boxHeight, _, _, _, _ = r.calcPatternDimesions(r.engine.livingCells)
	r.boxWidth += 4
	r.boxHeight += 4

	x := (r.gridWidth - r.boxWidth) / 2
	y := (r.gridHeight - r.boxHeight) / 2

	// Top and bottom borders.
	for col := x; col < x+r.boxWidth; col++ {
		r.screen.SetContent(
			col,
			y,
			tcell.RuneHLine,
			nil,
			css.box,
		)

		r.screen.SetContent(
			col,
			y+r.boxHeight-1,
			tcell.RuneHLine,
			nil,
			css.box,
		)
	}

	// Left and right borders.
	for row := y; row < y+r.boxHeight; row++ {
		r.screen.SetContent(
			x,
			row,
			tcell.RuneVLine,
			nil,
			css.box,
		)

		r.screen.SetContent(
			x+r.boxWidth-1,
			row,
			tcell.RuneVLine,
			nil,
			css.box,
		)
	}

	// Corners.
	r.screen.SetContent(
		x,
		y,
		tcell.RuneULCorner,
		nil,
		css.box,
	)

	r.screen.SetContent(
		x+r.boxWidth-1,
		y,
		tcell.RuneURCorner,
		nil,
		css.box,
	)

	r.screen.SetContent(
		x,
		y+r.boxHeight-1,
		tcell.RuneLLCorner,
		nil,
		css.box,
	)

	r.screen.SetContent(
		x+r.boxWidth-1,
		y+r.boxHeight-1,
		tcell.RuneLRCorner,
		nil,
		css.box,
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
