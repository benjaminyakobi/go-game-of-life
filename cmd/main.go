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
	def:            tcell.StyleDefault.Foreground(color.White).Background(color.Default),
	grey:           tcell.StyleDefault.Foreground(color.Black).Background(color.Grey),
	lightSlateGrey: tcell.StyleDefault.Foreground(color.Black).Background(color.LightSlateGrey),
	greenYellow:    tcell.StyleDefault.Foreground(color.Black).Background(color.GreenYellow),
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

var livingCells = make(LivingCellsSet)

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

func updateCellStyle(s tcell.Screen, w, h int) {
	if w%2 == 0 && h%2 != 0 {
		s.Put(w, h, ".", cellStyles.grey)
	} else if w%2 != 0 && h%2 == 0 {
		s.Put(w, h, ".", cellStyles.grey)
	} else {
		s.Put(w, h, ".", cellStyles.lightSlateGrey)
	}
}

func drawNewGrid(s tcell.Screen) {
	drawText(s, 0, 0, "Single click to select | Double click to unselect | Ctrl+R to run | Ctrl+C to exit")
	width, height := s.Size()
	for w := 0; w < width; w++ {
		for h := 2; h < height; h++ {
			updateCellStyle(s, w, h)
		}
	}
}

func runGameOfLife(s tcell.Screen) {
	q := make(chan LivingCell)
	ctx := context.Background()
	go setLivingCell(s, q, ctx) // consumer

	// producer
	go func() {
		s.DisableMouse()      // disabling mouse before running the game
		defer s.EnableMouse() // enabling mouse before returning
		i := 2                // TODO dummy value that should be replaced
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			// TODO implement game logic
			// TODO remove dummy i variable
			// TODO add new func that calculates the next generation
			// TODO implement by iterating over livingCells map (apply Life's rules)
			q <- LivingCell{posX: i, posY: i} // TODO dummy call that should be replaced
			i++
		}
	}()
}

func setLivingCell(s tcell.Screen, q <-chan LivingCell, ctx context.Context) {
	for {
		select {
		case lc, ok := <-q:
			if !ok {
				return
			}
			livingCells.Add(lc)
			drawText(s, 0, 1, fmt.Sprintf("cell [%v, %v] - living cells: %v", lc.posX, lc.posY, livingCells.Len()))
			s.Put(lc.posX, lc.posY, ".", cellStyles.greenYellow)
			s.Show()
		case <-ctx.Done():
			return
		}
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
	)

	// Event loop
	for {
		s.Show()           // Updating screen
		ev := <-s.EventQ() // Polling event

		// Processing the events
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			} else if ev.Key() == tcell.KeyCtrlL {
				s.Sync()
			} else if ev.Key() == tcell.KeyCtrlR {
				runGameOfLife(s)
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
				} else if y > 1 { // single-click
					livingCells.Add(LivingCell{posX: x, posY: y})
					drawText(s, 0, 1, fmt.Sprintf("selected [%v, %v] - living cells: %v", x, y, livingCells.Len()))
					s.Put(x, y, ".", cellStyles.greenYellow)
				}
				lastClickTime = now
				lastX, lastY = x, y
			}
		}
	}
}
