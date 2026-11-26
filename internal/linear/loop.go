package linear

import (
	"log"
	"time"
)

type loopState struct {
	lastClickTime time.Time
	lastKeyTime   time.Time // currently not in use
	lastX         int
	lastY         int

	dblClickDelay time.Duration

	running  bool
	boxOpen  bool
	boxIndex int

	interval time.Duration
	ticker   *time.Ticker
}

func initLoopState() *loopState {
	interval := 100 * time.Millisecond

	return &loopState{
		dblClickDelay: 500 * time.Millisecond,
		interval:      interval,
		ticker:        time.NewTicker(interval),
		boxIndex:      -1,
	}
}

func Start() {
	screen, err := initScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	cfg := loadConfig()

	w, _ := screen.Size()
	engine, err := newEngine(cfg, w)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	renderer, err := newRenderer(engine, screen)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer renderer.cleanup()

	renderer.initGame()

	state := initLoopState()
	defer state.ticker.Stop()

	renderer.runGameLoop(state)
}
