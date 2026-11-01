package onedgame

import (
	"log"
	"time"
)

type loopState struct {
	lastClickTime time.Time
	lastKeyTime   time.Time
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
	interval := 300 * time.Millisecond

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

	bitEngine, err := newBitEngine(screen.Size())
	if err != nil {
		log.Fatalf("%+v", err)
	}

	bitRenderer, err := newBitRenderer(bitEngine, screen)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer bitRenderer.cleanup()

	bitRenderer.initGame()

	state := initLoopState()
	defer state.ticker.Stop()

	bitRenderer.runGameLoop(state)
}
