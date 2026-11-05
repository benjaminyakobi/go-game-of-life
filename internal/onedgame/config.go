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
