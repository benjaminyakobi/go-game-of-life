package bitwise

import (
	"github.com/gdamore/tcell/v3"
)

type gridStyle struct {
	deadGlyph string
	liveGlyph string
	defStyle  tcell.Style
}
