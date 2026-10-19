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
	// br.drawGrid()
	// br.drawLiving(br.engine.getLiving())
	// br.screen.Show()
}
