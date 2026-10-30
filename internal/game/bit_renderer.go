package game

import (
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
	engine BitEngine
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

func newBitRenderer(be BitEngine, s tcell.Screen) (*bitRenderer, error) {
	s.EnableMouse()
	s.Clear()

	return &bitRenderer{
		screen: s,
		engine: be,
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
	br.drawLiving(br.engine.getLiving())
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

	br.removeLiving(br.engine.getLiving())
	br.engine.nextGen()

	living := br.engine.getLiving()
	br.drawLiving(living)

	if len(living) == 0 {
		state.running = false
		br.screen.EnableMouse()
		br.engine.resetGen()
	}
	br.screen.Show()
}

func (br *bitRenderer) eventDispatcher(ls *loopState, ev tcell.Event) bool {
	// TODO: add new events: select/unselect, run/stop
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			return true
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

func (br *bitRenderer) removeLiving(living []int) {
	for _, v := range living {
		x, y := br.engine.revertIndex(v)
		br.drawDead(x, y)
	}
}

func (br *bitRenderer) drawLiving(living []int) {
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
