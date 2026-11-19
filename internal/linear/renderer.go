package onedgame

import (
	"math"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type gridStyle struct {
	deadGlyph string
	liveGlyph string
	defStyle  tcell.Style
}

var gs = gridStyle{
	deadGlyph: ".",
	liveGlyph: "◆",
	defStyle: tcell.StyleDefault.
		Background(color.Black).
		Foreground(color.White),
}

type bitRenderer struct {
	screen tcell.Screen
	engine Engine
}

func initScreen() (tcell.Screen, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	return screen, nil
}

func newBitRenderer(e Engine, s tcell.Screen) (*bitRenderer, error) {
	s.EnableMouse()
	s.Clear()

	return &bitRenderer{
		screen: s,
		engine: e,
	}, nil
}

func (br *bitRenderer) cleanup() {
	maybePanic := recover()

	br.screen.Fini()

	if maybePanic != nil {
		panic(maybePanic)
	}
}

func (br *bitRenderer) initGame() {
	br.drawGrid()
	br.drawLivingCells(br.engine.getLiving())
	br.screen.Show()
}

func (br *bitRenderer) drawGrid() {
	width, height := br.screen.Size()

	for w := range width {
		for h := range height {
			br.drawDead(w, h)
		}
	}
}

func (br *bitRenderer) runGameLoop(state *loopState) {
	for {
		select {
		case <-state.ticker.C:
			br.handleTick(state)

		case ev := <-br.screen.EventQ():
			if br.eventDispatcher(state, ev) {
				return
			}
		}
	}
}

func (br *bitRenderer) handleTick(state *loopState) {
	if !state.running {
		return
	}

	deadCells, livingCells := br.engine.nextGen()
	br.drawDeadCells(deadCells)
	br.drawLivingCells(livingCells)

	if len(livingCells) == 0 {
		state.running = false
		br.screen.EnableMouse()
		br.engine.resetGen()
	}
	br.screen.Show()
}

func (br *bitRenderer) eventDispatcher(ls *loopState, ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			return true
		} else if ev.Key() == tcell.KeyRune && ev.Str() == "r" {
			if !ls.running {
				ls.running = true
				br.screen.DisableMouse()
			}
		} else if ev.Key() == tcell.KeyRune && ev.Str() == "s" {
			if !ls.running {
				return false
			}
			ls.running = !ls.running
			br.screen.EnableMouse()
		} else if ev.Key() == tcell.KeyRune && ev.Str() == "b" {
			if ls.running {
				return false
			}
			ls.boxIndex++
			br.engine.resetGen()
			for _, c := range br.engine.getLiving() {
				x, y := br.engine.revertIndex(c)
				br.engine.removeLiving(c)
				br.drawDead(x, y)
			}
			_, pattern := br.engine.getPattern(ls.boxIndex)
			w, h := br.screen.Size()
			centeredPattern := br.centerCells(
				pattern,
				0, 0, w, h,
			)
			br.engine.initLiving(centeredPattern)
			br.drawLivingCells(br.engine.getLiving())
			br.screen.Show()
		}
	case *tcell.EventMouse:
		if ev.Buttons() == tcell.ButtonPrimary {
			now := time.Now()
			x, y := ev.Position()
			idx := br.engine.createIndex(x, y)

			if now.Sub(ls.lastClickTime) <= ls.dblClickDelay &&
				(x == ls.lastX && y == ls.lastY) {
				// NOTE: 2d index -> 1d index -> remove cell
				br.engine.removeLiving(idx)
				br.drawDead(x, y)

			} else {
				// NOTE: 2d index -> 1d index -> draw cell
				br.engine.setLiving(idx)
				br.drawLive(x, y)
			}
			br.screen.Show()
			ls.lastClickTime = now
			ls.lastX = x
			ls.lastY = y
		}
	}
	return false
}

func (br *bitRenderer) drawDeadCells(living []int) {
	for _, v := range living {
		x, y := br.engine.revertIndex(v)
		br.drawDead(x, y)
	}
}

func (br *bitRenderer) drawLivingCells(living []int) {
	for _, v := range living {
		x, y := br.engine.revertIndex(v)
		br.drawLive(x, y)
	}
}

func (br *bitRenderer) drawDead(w, h int) {
	br.screen.Put(w, h, gs.deadGlyph, gs.defStyle)
}

func (br *bitRenderer) drawLive(w, h int) {
	br.screen.Put(w, h, gs.liveGlyph, gs.defStyle)
}

func (br *bitRenderer) calcPatternDimesions(fcs flatCells) (
	int, int, int, int, int, int) {
	minX := math.MaxInt
	minY := math.MaxInt
	maxX := math.MinInt
	maxY := math.MinInt

	for _, fc := range fcs {
		x, y := br.engine.revertIndex(fc)
		minX = min(minX, x)
		minY = min(minY, y)
		maxX = max(maxX, x)
		maxY = max(maxY, y)
	}

	patternWidth := maxX - minX + 1
	patternHeight := maxY - minY + 1

	return patternWidth, patternHeight, minX, minY, maxX, maxY
}

func (br *bitRenderer) centerCells(
	fcs flatCells, x, y, width, height int) flatCells {
	if len(fcs) == 0 {
		return fcs
	}

	patternWidth, patternHeight, minX, minY, _, _ :=
		br.calcPatternDimesions(fcs)

	offsetX := x + (width-patternWidth)/2
	offsetY := y + (height-patternHeight)/2

	var centeredFCS flatCells

	for _, fc := range fcs {
		x1, y1 := br.engine.revertIndex(fc)
		cfc := br.engine.createIndex(
			offsetX+(x1-minX),
			offsetY+(y1-minY),
		)
		centeredFCS = append(centeredFCS, cfc)
	}

	return centeredFCS
}
