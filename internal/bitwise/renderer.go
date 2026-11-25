package bitwise

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

func (r *renderer) drawLive(w, h int) {
	r.screen.Put(w, h, gs.liveGlyph, gs.defStyle)
}

func (r *renderer) drawDead(w, h int) {
	r.screen.Put(w, h, gs.deadGlyph, gs.defStyle)
}
