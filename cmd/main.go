package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
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

type LivingCell struct {
	posX int
	posY int
}

type LivingCellsSet map[LivingCell]struct{}

func (lcs LivingCellsSet) Add(lc LivingCell) {
	lcs[lc] = struct{}{}
}

func (lcs LivingCellsSet) Remove(lc LivingCell) {
	delete(lcs, lc)
}

func (lcs LivingCellsSet) Contains(lc LivingCell) bool {
	_, ok := lcs[lc]
	return ok
}

func (lcs LivingCellsSet) Len() int {
	return len(lcs)
}

const screenOffset = 2

var livingCells = make(LivingCellsSet)
var generation = 0
var ctx context.Context
var cancel context.CancelFunc
var gameIsRuning = false
var directions = [][]int{
	{-1, -1}, // top left
	{0, -1},  // top mid
	{1, -1},  // top right
	{-1, 0},  // left
	{1, 0},   // right
	{-1, 1},  // bottom left
	{0, 1},   // bottom mid
	{1, 1},   // bottom right
}

func clearLine(s tcell.Screen, y int) {
	w, _ := s.Size()
	for x := 0; x < w; x++ {
		s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
	}
}

func drawText(s tcell.Screen, x, y int, text string) {
	clearLine(s, y)
	col := x
	row := y
	var width int
	for text != "" {
		text, width = s.Put(col, row, text, cellStyles.def)
		col += width
		if width == 0 {
			// incomplete grapheme at end of string
			break
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

	if x == 0 && y == 2 {
		s.Put(x, y, string(tcell.RuneULCorner), cellStyles.def)
	} else if x == w-1 && y == 2 {
		s.Put(x, y, string(tcell.RuneURCorner), cellStyles.def)
	} else if x == 0 && y == h-1 {
		s.Put(x, y, string(tcell.RuneLLCorner), cellStyles.def)
	} else if x == w-1 && y == h-1 {
		s.Put(x, y, string(tcell.RuneLRCorner), cellStyles.def)
	}
}

func drawNewGrid(s tcell.Screen) {
	drawText(s, 0, 0, "Click select | Double click unselect | Ctrl+R run | Ctrl+P pause | Ctrl+C exit")
	width, height := s.Size()
	for w := 0; w < width; w++ {
		for h := screenOffset; h < height; h++ {
			updateCellStyle(s, w, h)
		}
	}
}

func runGameOfLife(s tcell.Screen, ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	s.DisableMouse() // disabling mouse before running the game
	defer ticker.Stop()
	defer s.EnableMouse() // enabling mouse before returning
	for {
		select {
		case <-ticker.C:
			calcNextGeneration(s)
			s.Show()
		case <-ctx.Done():
			drawText(s, 0, 1, fmt.Sprintf("stopped after %v generations", generation))
			s.Show()
			gameIsRuning = false
			generation = 0
			return
		}
	}
}

func calcNextGenDeadCells(s tcell.Screen, lc LivingCell) bool {
	count := 0
	w, h := s.Size()
	for _, d := range directions {
		dx, dy := d[0], d[1]
		if livingCells.Contains(LivingCell{posX: lc.posX + dx, posY: lc.posY + dy}) {
			count++
		}
		if count > 3 {
			return false
		}
	}
	if count == 3 {
		if lc.posY > screenOffset && lc.posY < h-1 && lc.posX > 0 && lc.posX < w-1 {
			s.Put(lc.posX, lc.posY, "@", cellStyles.greenYellow)
		}
		return true
	}
	return false
}

func calcNextGeneration(s tcell.Screen) {
	w, h := s.Size()
	livingCellsNextGen := make(LivingCellsSet)
	for lc := range livingCells {
		count := 0
		for _, d := range directions {
			dx, dy := d[0], d[1]
			neighborCell := LivingCell{posX: lc.posX + dx, posY: lc.posY + dy}
			if livingCells.Contains(neighborCell) {
				count++
			} else {
				ok := calcNextGenDeadCells(s, neighborCell)
				if ok {
					livingCellsNextGen.Add(neighborCell)
				}
			}
		}
		if count < 2 || count > 3 {
			if lc.posY > screenOffset && lc.posY < h-1 && lc.posX > 0 && lc.posX < w-1 {
				updateCellStyle(s, lc.posX, lc.posY)
			}
		} else if count == 2 || count == 3 {
			livingCellsNextGen.Add(lc)
		}
	}
	livingCells = livingCellsNextGen
	generation++
	drawText(s, 0, 1, fmt.Sprintf("generation: %v, living cells: %v", generation, livingCells.Len()))
	if livingCells.Len() == 0 {
		cancel()
	}
}

func main() {
	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(cellStyles.def)
	s.EnableMouse()
	s.Clear()

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	drawNewGrid(s)

	var (
		lastClickTime time.Time
		lastX, lastY  int
		dblClickDelay = 500 * time.Millisecond
		w, h          = s.Size()
	)

	// Event loop
	for {
		s.Show()           // Updating screen
		ev := <-s.EventQ() // Polling event

		// Processing the events
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h = s.Size()
			drawNewGrid(s)
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			} else if ev.Key() == tcell.KeyCtrlS {
				s.Sync()
			} else if ev.Key() == tcell.KeyCtrlR {
				if livingCells.Len() == 0 {
					drawText(s, 0, 1, fmt.Sprintf("not starting, select cells first %v", livingCells.Len()))
				} else {
					gameIsRuning = true
					ctx, cancel = context.WithCancel(context.Background())
					go runGameOfLife(s, ctx)
				}
			} else if ev.Key() == tcell.KeyCtrlP && gameIsRuning {
				cancel()
			}
		case *tcell.EventMouse:
			x, y := ev.Position()
			switch ev.Buttons() {
			case tcell.ButtonPrimary:
				now := time.Now()
				if now.Sub(lastClickTime) <= dblClickDelay &&
					x == lastX && y == lastY { // double-click
					livingCells.Remove(LivingCell{posX: x, posY: y})
					drawText(s, 0, 1, fmt.Sprintf("unselected [%v, %v] - living cells: %v", x, y, livingCells.Len()))
					updateCellStyle(s, x, y)
				} else if y > screenOffset && y < h-1 && x > 0 && x < w-1 { // single-click
					livingCells.Add(LivingCell{posX: x, posY: y})
					drawText(s, 0, 1, fmt.Sprintf("selected [%v, %v] - living cells: %v", x, y, livingCells.Len()))
					s.Put(x, y, "@", cellStyles.greenYellow)
				}
				lastClickTime = now
				lastX, lastY = x, y
			}
		}
	}
}
