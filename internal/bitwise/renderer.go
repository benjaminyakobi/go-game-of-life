package bitwise

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type gridStyle struct {
	deadGlyph string
	liveGlyph string
	defStyle  tcell.Style
}

var gs = gridStyle{
	deadGlyph: ".",
	liveGlyph: "◆",
	defStyle: tcell.StyleDefault.
		Background(color.Black).
		Foreground(color.White),
}
