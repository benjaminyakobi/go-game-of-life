package bitwise

import "log"

type loopState struct {
}

func newLoopState() (*loopState, error) {
	return &loopState{}, nil
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
