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
	interval := 500 * time.Millisecond

	return &loopState{
		dblClickDelay: 500 * time.Millisecond,
		interval:      interval,
		ticker:        time.NewTicker(interval),
	}
}

func cleanup(renderer *renderer) {
	maybePanic := recover()

	renderer.screen.Fini()

	if maybePanic != nil {
		panic(maybePanic)
	}
}

func handleTick(renderer *renderer, engine *engine, state *loopState) {
	if !state.running {
		return
	}

	renderer.killLivingCellsOnGrid(engine.livingCells)

	engine.calcNextGeneration()

	renderer.drawLivingCellsOnGrid(engine.livingCells)
	renderer.drawDeadCellsOnGrid(engine.deadCells)

	var text string

	if engine.livingCells.Len() == 0 {
		state.running = false
		renderer.screen.EnableMouse()

		text = fmt.Sprintf(
			"stopped after %v generations",
			engine.generation,
		)
	} else {
		text = fmt.Sprintf(
			"generation: %v, living cells: %v",
			engine.generation,
			engine.livingCells.Len(),
		)
	}

	renderer.drawText(1, text)
	renderer.screen.Show()
}

func initializeGame(renderer *renderer, engine *engine) {
	renderer.drawNewGrid()
	renderer.drawLivingCellsOnGrid(engine.livingCells)
	renderer.screen.Show()
}

func runGameLoop(renderer *renderer, engine *engine, state *loopState) {
	for {
		select {
		case <-state.ticker.C:
			handleTick(renderer, engine, state)

		case ev := <-renderer.screen.EventQ():
			if handleEvent(renderer, engine, state, ev) {
				return
			}
		}
	}
}

func Start() {
	cfg := loadConfig()

	engine := initEngine(cfg)

	renderer, err := initRenderer(engine)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	defer cleanup(renderer)

	initializeGame(renderer, engine)

	state := initLoopState()
	defer state.ticker.Stop()

	runGameLoop(renderer, engine, state)
}
