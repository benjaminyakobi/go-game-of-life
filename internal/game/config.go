package game

// Contains game config code

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

type indexedPatterns struct {
	order []string
	sets  map[string]cellsSet
}

func (ip *indexedPatterns) Len() int {
	return len(ip.order)
}

func (ip *indexedPatterns) Get(i int) (string, cellsSet) {
	cyclicIndex := i % len(ip.order)
	patternName := ip.order[cyclicIndex]
	return patternName, ip.sets[patternName]
}

func (c *config) loadPatterns() *indexedPatterns {
	out := make(map[string]cellsSet, len(c.Patterns))
	for name, cells := range c.Patterns {
		set := make(cellsSet, len(cells))
		for _, c := range cells {
			set[c] = struct{}{}
		}
		out[name] = set
	}

	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}

	return &indexedPatterns{order: keys, sets: out}
}

func loadConfig() config {
	file, err := os.Open("./config.json")
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
