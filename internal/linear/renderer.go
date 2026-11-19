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

type renderer struct {
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

func newRenderer(e Engine, s tcell.Screen) (*renderer, error) {
	s.EnableMouse()
	s.Clear()

	return &renderer{
		screen: s,
		engine: e,
	}, nil
}

func (r *renderer) cleanup() {
	maybePanic := recover()

	r.screen.Fini()

	if maybePanic != nil {
		panic(maybePanic)
	}
}

func (r *renderer) initGame() {
	r.drawGrid()
	r.drawLivingCells(r.engine.getLiving())
	r.screen.Show()
}

func (r *renderer) drawGrid() {
	width, height := r.screen.Size()

	for w := range width {
		for h := range height {
			r.drawDead(w, h)
		}
	}
}

func (r *renderer) runGameLoop(state *loopState) {
	for {
		select {
		case <-state.ticker.C:
			r.handleTick(state)

		case ev := <-r.screen.EventQ():
			if r.eventDispatcher(state, ev) {
				return
			}
		}
	}
}

func (r *renderer) handleTick(state *loopState) {
	if !state.running {
		return
	}

	deadCells, livingCells := r.engine.nextGen()
	r.drawDeadCells(deadCells)
	r.drawLivingCells(livingCells)

	if len(livingCells) == 0 {
		state.running = false
		r.screen.EnableMouse()
		r.engine.resetGen()
	}
	r.screen.Show()
}

func (r *renderer) eventDispatcher(ls *loopState, ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			return true
		} else if ev.Key() == tcell.KeyRune && ev.Str() == "r" {
			if !ls.running {
				ls.running = true
				r.screen.DisableMouse()
			}
		} else if ev.Key() == tcell.KeyRune && ev.Str() == "s" {
			if !ls.running {
				return false
			}
			ls.running = !ls.running
			r.screen.EnableMouse()
		} else if ev.Key() == tcell.KeyRune && ev.Str() == "b" {
			if ls.running {
				return false
			}
			ls.boxIndex++
			r.engine.resetGen()
			for _, c := range r.engine.getLiving() {
				x, y := r.engine.revertIndex(c)
				r.engine.removeLiving(c)
				r.drawDead(x, y)
			}
			_, pattern := r.engine.getPattern(ls.boxIndex)
			w, h := r.screen.Size()
			centeredPattern := r.centerCells(
				pattern,
				0, 0, w, h,
			)
			r.engine.initLiving(centeredPattern)
			r.drawLivingCells(r.engine.getLiving())
			r.screen.Show()
		}
	case *tcell.EventMouse:
		if ev.Buttons() == tcell.ButtonPrimary {
			now := time.Now()
			x, y := ev.Position()
			idx := r.engine.createIndex(x, y)

			if now.Sub(ls.lastClickTime) <= ls.dblClickDelay &&
				(x == ls.lastX && y == ls.lastY) {
				// NOTE: 2d index -> 1d index -> remove cell
				r.engine.removeLiving(idx)
				r.drawDead(x, y)

			} else {
				// NOTE: 2d index -> 1d index -> draw cell
				r.engine.setLiving(idx)
				r.drawLive(x, y)
			}
			r.screen.Show()
			ls.lastClickTime = now
			ls.lastX = x
			ls.lastY = y
		}
	}
	return false
}

func (r *renderer) drawDeadCells(living []int) {
	for _, v := range living {
		x, y := r.engine.revertIndex(v)
		r.drawDead(x, y)
	}
}

func (r *renderer) drawLivingCells(living []int) {
	for _, v := range living {
		x, y := r.engine.revertIndex(v)
		r.drawLive(x, y)
	}
}

func (r *renderer) drawDead(w, h int) {
	r.screen.Put(w, h, gs.deadGlyph, gs.defStyle)
}

func (r *renderer) drawLive(w, h int) {
	r.screen.Put(w, h, gs.liveGlyph, gs.defStyle)
}

func (r *renderer) calcPatternDimesions(fcs flatCells) (
	int, int, int, int, int, int) {
	minX := math.MaxInt
	minY := math.MaxInt
	maxX := math.MinInt
	maxY := math.MinInt

	for _, fc := range fcs {
		x, y := r.engine.revertIndex(fc)
		minX = min(minX, x)
		minY = min(minY, y)
		maxX = max(maxX, x)
		maxY = max(maxY, y)
	}

	patternWidth := maxX - minX + 1
	patternHeight := maxY - minY + 1

	return patternWidth, patternHeight, minX, minY, maxX, maxY
}

func (r *renderer) centerCells(
	fcs flatCells, x, y, width, height int) flatCells {
	if len(fcs) == 0 {
		return fcs
	}

	patternWidth, patternHeight, minX, minY, _, _ :=
		r.calcPatternDimesions(fcs)

	offsetX := x + (width-patternWidth)/2
	offsetY := y + (height-patternHeight)/2

	var centeredFCS flatCells

	for _, fc := range fcs {
		x1, y1 := r.engine.revertIndex(fc)
		cfc := r.engine.createIndex(
			offsetX+(x1-minX),
			offsetY+(y1-minY),
		)
		centeredFCS = append(centeredFCS, cfc)
	}

	return centeredFCS
}
