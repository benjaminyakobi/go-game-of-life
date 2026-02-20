package main

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

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

type LivingCell struct {
	PosX int `json:"x"`
	PosY int `json:"y"`
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

func (lcs LivingCellsSet) Copy() LivingCellsSet {
	if lcs == nil {
		return nil
	}
	lcsCopy := make(LivingCellsSet, lcs.Len())
	for cell := range lcs {
		lcsCopy[cell] = struct{}{}
	}
	return lcsCopy
}

const screenOffset = 1
const historySize = 50

type Config struct {
	Patterns map[string][]LivingCell `json:"patterns"`
}

var m time.Duration = 500
var predefinedLivingCells []LivingCellsSet
var historyLivingCells = list.New()
var boxWidth, boxHeight, minWidth, minHeight = -1, -1, math.MaxInt32, math.MinInt32
var predefinedLCIndex = 0
var boxOpen = false
var gameText = ""
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

func loadConfig() {
	file, err := os.Open("./conf.json")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var conf Config
	if err := decoder.Decode(&conf); err != nil {
		if err == io.EOF {
			fmt.Println("finished decoding config file")
		} else {
			log.Fatalf("failed to open file: %v", err)
		}
	}

	for _, points := range conf.Patterns {
		var lcs = make(LivingCellsSet)
		for i := 0; i < len(points); i++ {
			lcs.Add(LivingCell{PosX: points[i].PosX, PosY: points[i].PosY})
		}
		predefinedLivingCells = append(predefinedLivingCells, lcs)
	}
}

func clearLine(s tcell.Screen, y int) {
	w, _ := s.Size()
	for x := 0; x < w; x++ {
		s.SetContent(x, y, ' ', nil, tcell.StyleDefault)
	}
}

func drawText(s tcell.Screen, y int, text string) {
	clearLine(s, y)
	w, h := s.Size()
	textWidth := runewidth.StringWidth(text)

	calcX := (w - textWidth) / 2

	for r := 0; r < calcX; r++ {
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
	for w := 0; w < width; w++ {
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

func runGameOfLife(s tcell.Screen, ctx context.Context, pauseChan <-chan bool, millisChan <-chan time.Duration) {
	ticker := time.NewTicker(m * time.Millisecond)
	s.DisableMouse() // disabling mouse before running the game
	defer ticker.Stop()
	defer s.EnableMouse() // enabling mouse before returning
	for {
		select {
		case <-ticker.C:
			calcNextGeneration(s)
			s.Show()
		case <-ctx.Done():
			gameText = fmt.Sprintf("stopped after %v generations", generation)
			drawText(s, 1, gameText)
			s.Show()
			gameIsRuning = false
			generation = 0
			return
		case <-pauseChan:
			gameText = fmt.Sprintf("paused after %v generations", generation)
			drawText(s, 1, gameText)
			s.Show()
			gameIsRuning = false
			return
		case <-millisChan:
			ticker = time.NewTicker(m * time.Millisecond)
		}
	}
}

func calcNextGenDeadCells(s tcell.Screen, lc LivingCell) bool {
	count := 0
	w, h := s.Size()
	for _, d := range directions {
		dx, dy := d[0], d[1]
		if livingCells.Contains(LivingCell{PosX: lc.PosX + dx, PosY: lc.PosY + dy}) {
			count++
		}
		if count > 3 {
			return false
		}
	}
	if count == 3 {
		if lc.PosY > screenOffset && lc.PosY < h-1 && lc.PosX > 0 && lc.PosX < w-1 {
			s.Put(lc.PosX, lc.PosY, "@", cellStyles.greenYellow)
		}
		return true
	}
	return false
}

func calcNextGeneration(s tcell.Screen) {
	w, h := s.Size()
	if historyLivingCells.Len() > historySize {
		historyLivingCells.Remove(historyLivingCells.Front())
	}
	historyLivingCells.PushBack(livingCells)
	livingCellsNextGen := make(LivingCellsSet)
	for lc := range livingCells {
		count := 0
		for _, d := range directions {
			dx, dy := d[0], d[1]
			neighborCell := LivingCell{PosX: lc.PosX + dx, PosY: lc.PosY + dy}
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
			if lc.PosY > screenOffset && lc.PosY < h-1 && lc.PosX > 0 && lc.PosX < w-1 {
				updateCellStyle(s, lc.PosX, lc.PosY)
			}
		} else if count == 2 || count == 3 {
			livingCellsNextGen.Add(lc)
		}
	}
	livingCells = livingCellsNextGen
	generation++
	gameText = fmt.Sprintf("generation: %v, living cells: %v", generation, livingCells.Len())
	drawText(s, 1, gameText)
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

	loadConfig()
	drawNewGrid(s)
	drawLivingCellsOnGrid(s)

	var (
		lastClickTime time.Time
		lastKeyTime   time.Time
		lastX, lastY  int
		dblClickDelay = 500 * time.Millisecond
		w, h          = s.Size()
		pauseChan     = make(chan bool)
		millisChan    = make(chan time.Duration)
	)

	// Event loop
	for {
		s.Show()           // Updating screen
		ev := <-s.EventQ() // Polling event

		// Processing the events
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h = s.Size()
			if boxOpen {
				drawNewGrid(s)
				drawBox(s, "Choose predefined pattern")
			} else {
				drawNewGrid(s)
				drawLivingCellsOnGrid(s)
			}
		case *tcell.EventKey:
			keyNow := time.Now()
			if ev.Key() == tcell.KeyEscape {
				return
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "p" && boxOpen == false && historyLivingCells.Len() > 0 {
				if gameIsRuning {
					pauseChan <- true
				}
			} else if ev.Key() == tcell.KeyLeft && !gameIsRuning && !boxOpen && historyLivingCells.Len() > 0 {
				historyVal := historyLivingCells.Back()
				historyLivingCells.Remove(historyVal)
				livingCells = historyVal.Value.(LivingCellsSet)
				if generation > 0 {
					generation--
				}
				gameText = fmt.Sprintf("history | generation: %v, living cells: %v", generation, livingCells.Len())
				drawNewGrid(s)
				drawLivingCellsOnGrid(s)
			} else if ev.Key() == tcell.KeyRight && !gameIsRuning && !boxOpen {
				calcNextGeneration(s)
				s.Show()
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "b" && !gameIsRuning && len(predefinedLivingCells) > 0 {
				s.DisableMouse()
				boxOpen = true
				drawNewGrid(s)
				livingCells = predefinedLivingCells[predefinedLCIndex%len(predefinedLivingCells)].Copy()
				drawBox(s, "Choose predefined pattern")
				predefinedLCIndex++
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "r" {
				if boxOpen {
					s.EnableMouse()
					drawNewGrid(s)
					drawLivingCellsOnGrid(s)
					drawText(s, 1, "Chosen predefined pattern")
					predefinedLCIndex--
					boxOpen = !boxOpen
				} else if livingCells.Len() == 0 {
					gameText = fmt.Sprintf("not starting, select cells first %v", livingCells.Len())
					drawText(s, 1, gameText)
				} else if !gameIsRuning {
					gameIsRuning = true
					ctx, cancel = context.WithCancel(context.Background())
					go runGameOfLife(s, ctx, pauseChan, millisChan)
				}
			} else if keyNow.Sub(lastKeyTime) <= dblClickDelay && ev.Key() == tcell.KeyRune && ev.Str() == "=" {
				if m > 100 && gameIsRuning {
					m -= 100
					millisChan <- m
				}
			} else if keyNow.Sub(lastKeyTime) <= dblClickDelay && ev.Key() == tcell.KeyRune && ev.Str() == "-" {
				if m < 1000 && gameIsRuning {
					m += 100
					millisChan <- m
				}
			} else if ev.Key() == tcell.KeyRune && ev.Str() == "s" && gameIsRuning {
				cancel()
			}
			lastKeyTime = keyNow
		case *tcell.EventMouse:
			x, y := ev.Position()
			switch ev.Buttons() {
			case tcell.ButtonPrimary:
				now := time.Now()
				if now.Sub(lastClickTime) <= dblClickDelay &&
					x == lastX && y == lastY && y > screenOffset && y < h-1 && x > 0 && x < w-1 { // double-click
					livingCells.Remove(LivingCell{PosX: x, PosY: y})
					gameText = fmt.Sprintf("unselected [%v, %v] - living cells: %v", x, y, livingCells.Len())
					drawText(s, 1, gameText)
					updateCellStyle(s, x, y)
				} else if y > screenOffset && y < h-1 && x > 0 && x < w-1 { // single-click
					livingCells.Add(LivingCell{PosX: x, PosY: y})
					gameText = fmt.Sprintf("selected [%v, %v] - living cells: %v", x, y, livingCells.Len())
					drawText(s, 1, gameText)
					s.Put(x, y, "@", cellStyles.greenYellow)
				}
				lastClickTime = now
				lastX, lastY = x, y
			}
		}
	}
}
