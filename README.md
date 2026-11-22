# go-game-of-life
> Go implementation of John Conway's Game Of Life that runs in the Terminal

[Conway's Game Of Life Simulation](https://playgameoflife.com)

[Backpressure Patterns in Go: From Channels to Queues to Load Shedding](https://medium.com/@Realblank/backpressure-patterns-in-go-from-channels-to-queues-to-load-shedding-0841c9fe5607)

## Rules
[Wikipedia - Conway's Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life)

Every cell interacts with its eight neighbours, which are the cells that are horizontally, vertically, or diagonally adjacent. At each step in time, the following transitions occur:
- Any live cell with fewer than two live neighbours dies, as if by underpopulation.
- Any live cell with two or three live neighbours lives on to the next generation.
- Any live cell with more than three live neighbours dies, as if by overpopulation.
- Any dead cell with exactly three live neighbours becomes a live cell, as if by reproduction.

## Key Bindings
- `Mouse Left Click` - Choose cell
- `Mouse Right Click` - Unchoose cell
- `b` - Clear Grid & Choose Predefined Patterns
- `r` - Run
- `p` - Pause
- `s` - Stop & Reset Generations
- `=` - Increase Speed
- `-` - Decrease Speed
- `Left Arrow` - When game is paused: Previous Generation
- `Right Arrow` - When game is paused: Next Generation

## Architecture

The application is split into small files under `internal/game/`, with each file owning one part of the terminal game:

- `config.go` loads `conf.json` and predefined patterns.
- `engine.go` owns the Game of Life state: living cells, dead cells, generation count, patterns, and history.
- `handlers.go` handles keyboard, mouse, and resize events from the terminal (when the amount of code is low the code is inside the `renderer.go` file).
- `loop.go` starts the game and runs the main loop using a ticker and terminal event queue.
- `renderer.go` draws the grid, cells, text, and pattern picker using `tcell`.

