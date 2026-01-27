package main

import (
	"log"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

func main() {
	defStyle := tcell.StyleDefault.Background(color.Reset).Foreground(color.Reset)

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.Clear()

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	// Draw grid
	cellStyleOne := tcell.StyleDefault.Foreground(color.Black).Background(color.Grey)
	cellStyleTwo := tcell.StyleDefault.Foreground(color.Black).Background(color.NewRGBColor(150, 150, 150))
	width, height := s.Size()
	for w := 0; w < width; w++ {
		for h := 0; h < height; h++ {
			if w%2 == 0 && h%2 != 0 {
				s.Put(w, h, ".", cellStyleOne)
			} else if w%2 != 0 && h%2 == 0 {
				s.Put(w, h, ".", cellStyleOne)
			} else {
				s.Put(w, h, ".", cellStyleTwo)
			}
		}
	}

	// Event loop
	for {
		s.Show()           // Updating screen
		ev := <-s.EventQ() // Polling event

		// Processing the events
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return
			} else if ev.Key() == tcell.KeyCtrlL {
				s.Sync()
			}
		}
	}
}
