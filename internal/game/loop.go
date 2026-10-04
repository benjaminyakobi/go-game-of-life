package game

// Contains game loop code

import (
	"fmt"
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

func cleanup(renderer *renderer) {
	maybePanic := recover()

	renderer.screen.Fini()

	if maybePanic != nil {
		panic(maybePanic)
	}
}

func handleTick(renderer *renderer, state *loopState) {
	if !state.running {
		return
	}

	renderer.engine.NextGeneration()

	renderer.drawDeadCellsOnGrid(renderer.engine.DeadCells())
	renderer.drawLivingCellsOnGrid(renderer.engine.GetLivingCells())

	var text string

	if renderer.engine.GetLivingCells().Len() == 0 {
		state.running = false
		renderer.screen.EnableMouse()

		text = fmt.Sprintf(
			"stopped after %v generations",
			renderer.engine.GetGeneration(),
		)
		renderer.engine.ResetGeneration()
	} else {
		text = fmt.Sprintf(
			"generation: %v, living cells: %v",
			renderer.engine.GetGeneration(),
			renderer.engine.GetLivingCells().Len(),
		)
	}

	renderer.drawText(1, text)
	renderer.screen.Show()
}

func initializeGame(renderer *renderer) {
	renderer.drawNewGrid()
	renderer.drawLivingCellsOnGrid(renderer.engine.GetLivingCells())
	renderer.screen.Show()
}

func runGameLoop(renderer *renderer, state *loopState) {
	for {
		select {
		case <-state.ticker.C:
			handleTick(renderer, state)

		case ev := <-renderer.screen.EventQ():
			if handleEvent(renderer, state, ev) {
				return
			}
		}
	}
}

func Start() {
	cfg := loadConfig()

	engine := newEngine(cfg)

	renderer, err := initRenderer(engine)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	defer cleanup(renderer)

	initializeGame(renderer)

	state := initLoopState()
	defer state.ticker.Stop()

	runGameLoop(renderer, state)
}
