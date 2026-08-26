package game

// Contains game common code

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"
)

const screenOffset = 1
const historySize = 50

type Config struct {
	Patterns map[string][]livingCell `json:"patterns"`
}

var m time.Duration = 500
var predefinedLivingCells []LivingCellsSet
var historyLivingCells = list.New()
var boxWidth, boxHeight, minWidth, minHeight = -1, -1, math.MaxInt32, math.MinInt32
var predefinedLCIndex = 0
var boxOpen = false
var gameText = ""
var livingCells = make(LivingCellsSet)
var generation = 0
var ctx context.Context
var cancel context.CancelFunc
var gameIsRuning = false
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
	var conf Config
	if err := decoder.Decode(&conf); err != nil {
		if err == io.EOF {
			fmt.Println("finished decoding config file")
		} else {
			log.Fatalf("failed to open file: %v", err)
		}
	}

	for _, points := range conf.Patterns {
		var lcs = make(LivingCellsSet)
		for i := range points {
			lcs.Add(livingCell{PosX: points[i].PosX, PosY: points[i].PosY})
		}
		predefinedLivingCells = append(predefinedLivingCells, lcs)
	}
}
