package game

import (
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
			br.screen.Put(w, h, gs.deadGlyph, gs.defStyle)
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

	br.engine.nextGen() // TODO: engine's func body not implemented yet

	living := br.engine.getLiving()
	br.drawLiving(living) // TODO: func body not implemented yet

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
	}
	return false
}

func (br *bitRenderer) drawLiving(living []int) {
	// TODO: implement
}

func (br *bitRenderer) createIndex(x, y int) int {
	w, _ := br.screen.Size()
	return y*w + x
}

func (br *bitRenderer) revertIndex(idx int) (int, int) {
	w, _ := br.screen.Size()
	return idx % w, idx / w
}
