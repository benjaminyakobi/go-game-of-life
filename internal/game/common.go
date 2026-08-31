package game

// Contains game common code

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

// TODO: should be removed from here
const screenOffset = 1
const historySize = 50

type config struct {
	Patterns map[string][]livingCell `json:"patterns"`
}

// TODO: move to engine.go / grid.go
var predefinedLivingCells []livingCellsSet

// TODO: should be moved to grid.go
var boxWidth, boxHeight, minWidth, minHeight = -1, -1, math.MaxInt32, math.MinInt32

// TODO: should be removed from here
var predefinedLCIndex = 0

// TODO: should be removed from here
var gameText = ""

// TODO: move to engine.go / grid.go
var directions = [][]int{
	{-1, -1}, // top left
	{0, -1},  // top mid
	{1, -1},  // top right
	{-1, 0},  // left
	{1, 0},   // right
	{-1, 1},  // bottom left
	{0, 1},   // bottom mid
	{1, 1},   // bottom right
}

func loadConfig() {
	file, err := os.Open("./conf.json")
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var conf config
	if err := decoder.Decode(&conf); err != nil {
		if err == io.EOF {
			fmt.Println("finished decoding config file")
		} else {
			log.Fatalf("failed to open file: %v", err)
		}
	}

	for _, points := range conf.Patterns {
		var lcs = make(livingCellsSet)
		for i := range points {
			lcs.Add(livingCell{PosX: points[i].PosX, PosY: points[i].PosY})
		}
		predefinedLivingCells = append(predefinedLivingCells, lcs)
	}
}
