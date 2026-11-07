package onedgame

// TODO: load the config.json file for the predefined patterns!

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

type cell struct {
	PosX int `json:"x"`
	PosY int `json:"y"`
}

type flatCells []int

type config struct {
	Patterns map[string][]cell `json:"patterns"`
}

type flatIndexedPatterns struct {
	keys []string
	sets map[string]flatCells
}

func (ip *flatIndexedPatterns) Len() int {
	return len(ip.keys)
}

func (ip *flatIndexedPatterns) Get(i int) (string, flatCells) {
	cyclicIndex := i % len(ip.keys)
	patternName := ip.keys[cyclicIndex]
	return patternName, ip.sets[patternName]
}

// func (c *config) loadPatterns(width int) *flatIndexedPatterns {
// }
//
// func loadConfig() config {
// }
