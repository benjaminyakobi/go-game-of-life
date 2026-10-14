package game

// Contains game loop code

import (
	"fmt"
	"log"
	"os"
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

func Start(engineType string) {
	switch engineType {
	case "default":
		cfg := loadConfig()

		engine := newEngine(cfg)

		renderer, err := initRenderer(engine)
		if err != nil {
			log.Fatalf("%+v", err)
		}

		defer renderer.cleanup()

		renderer.initializeGame()

		state := initLoopState()
		defer state.ticker.Stop()

		renderer.runGameLoop(state)
	case "bit":
		// TODO: implement
	default:
		fmt.Fprintln(os.Stderr, "unknown engine:", engineType)
		os.Exit(1)
	}
}
