package game

// Contains game common code

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type config struct {
	Patterns map[string][]cell `json:"patterns"`
}

// TODO: should be removed from here
var predefinedLCIndex = 0

func (c *config) loadPatterns() []cellsSet {
	var patterns []cellsSet
	for _, points := range c.Patterns {
		var cs = make(cellsSet)
		for i := range points {
			cs.Add(cell{PosX: points[i].PosX, PosY: points[i].PosY})
		}
		patterns = append(patterns, cs)
	}
	return patterns
}

func loadConfig() config {
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

	return conf
}
