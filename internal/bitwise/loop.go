package bitwise

import (
	"log"
	"time"
)

type loopState struct {
	lastClickTime time.Time
	dblClickDelay time.Duration
	lastX         int
	lastY         int
}

func newLoopState() (*loopState, error) {
	return &loopState{
		dblClickDelay: 500 * time.Millisecond,
	}, nil
}

func Start() {
	screen, err := initScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	w, h := screen.Size()
	engine, err := newEngine(w, h)

	renderer, err := newRenderer(engine, screen)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer renderer.cleanup()

	renderer.initGame()

	ls, err := newLoopState()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	renderer.runGameLoop(ls)
}
